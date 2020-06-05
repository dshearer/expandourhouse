package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"io"
	"log"
	"strconv"
	"strings"

	"expandourhouse.com/mapdata/bulkInserter"
	"expandourhouse.com/mapdata/congresses"
	"expandourhouse.com/mapdata/housedb"
	"expandourhouse.com/mapdata/states"
	"github.com/pkg/errors"
)

const gSourceName = "tufts-turnout"

var gWordsToNumbers = map[string]int{
	"one":          1,
	"two":          2,
	"three":        3,
	"four":         4,
	"five":         5,
	"six":          6,
	"seven":        7,
	"eight":        8,
	"nine":         9,
	"ten":          10,
	"eleven":       11,
	"twelve":       12,
	"thirteen":     13,
	"fourteen":     14,
	"fifteen":      15,
	"sixteen":      16,
	"seventeen":    17,
	"eighteen":     18,
	"nineteen":     19,
	"twenty":       20,
	"twenty one":   21,
	"twenty-one":   21,
	"twenty two":   22,
	"twenty three": 23,
	"twenty four":  24,
	"twenty five":  25,
	"twenty six":   26,
	"twenty seven": 27,
	"twenty eight": 28,
	"twenty nine":  29,
	"thirty":       30,
}

type turnoutRec struct {
	stateUsps string
	district  int
	year      int
	numVotes  int
}

func (self *turnoutRec) init(stateUsps string, district, year int) {
	self.stateUsps = stateUsps
	self.district = district
	self.year = year
	self.numVotes = 0
}

type readerAction int

const (
	gReaderActionReturnRec  readerAction = iota
	gReaderActionAbandonRec readerAction = iota
	gReaderActionAgain      readerAction = iota
)

var gPlaceCols = []string{"City", "County", "District", "Town", "Township",
	"Ward", "Parish", "Populated Place", "Hundred", "Borough"}

func valIsNull(val string) bool {
	return strings.ToLower(val) == "null" || len(strings.TrimSpace(val)) == 0
}

type turnoutReaderGroupById struct {
	csvReader    *csv.Reader
	colNameToIdx map[string]int

	// context:
	recBuff  []string
	recGroup [][]string
}

func newTurnoutReaderGroupById(f io.Reader) *turnoutReaderGroupById {
	r := &turnoutReaderGroupById{csv.NewReader(f), nil, nil, nil}
	r.csvReader.ReuseRecord = false
	r.csvReader.Comma = '\t'
	return r
}

func (self *turnoutReaderGroupById) read() ([][]string, error) {
	for {
		var action readerAction
		var err error
		if len(self.recGroup) == 0 {
			action, err = self.read_noCurrId()
		} else {
			action, err = self.read_currId()
		}
		if err != nil {
			return nil, err
		}

		switch action {
		case gReaderActionReturnRec:
			group := self.recGroup
			self.recGroup = nil
			return group, nil

		case gReaderActionAbandonRec:
			self.recGroup = nil

		case gReaderActionAgain:
		} // switch
	} // for
}

func (self *turnoutReaderGroupById) read_noCurrId() (readerAction, error) {
	// get next line
	rec, err := self.readNextLine()
	if err != nil {
		return gReaderActionAbandonRec, err
	}
	if len(rec) == 0 {
		return gReaderActionAgain, nil
	}

	if self.colNameToIdx == nil {
		// keep column names
		self.colNameToIdx = make(map[string]int)
		for idx, colName := range rec {
			colName = strings.ToLower(strings.TrimSpace(colName))
			self.colNameToIdx[colName] = idx
		}
		return gReaderActionAgain, nil
	}

	self.recGroup = [][]string{rec}
	return gReaderActionAgain, nil
}

func (self *turnoutReaderGroupById) read_currId() (readerAction, error) {
	idIdx := self.colNameToIdx["id"]
	currId := self.recGroup[0][idIdx]

	// get next line
	rec, err := self.readNextLine()
	if err != nil {
		return gReaderActionAbandonRec, err
	}
	if len(rec) == 0 {
		return gReaderActionAgain, nil
	}

	if rec[idIdx] == currId {
		self.recGroup = append(self.recGroup, rec)
		return gReaderActionAgain, nil
	} else {
		self.recBuff = rec
		return gReaderActionReturnRec, nil
	}
}

func (self *turnoutReaderGroupById) readNextLine() ([]string, error) {
	if self.recBuff != nil {
		rec := self.recBuff
		self.recBuff = nil
		return rec, nil
	}

	return self.csvReader.Read()
}

type turnoutReader struct {
	groupReader *turnoutReaderGroupById
}

func newTurnoutReader(f io.Reader) *turnoutReader {
	return &turnoutReader{newTurnoutReaderGroupById(f)}
}

func (self *turnoutReader) read() (*turnoutRec, error) {
	/*
		ID formats:

		al.congress.1819
			- No districts

		al.uscongress.1.1823
			- For 1st district

		ct.special.congress.1790
			- Special election

		ct.special2.congress.1793
			- Special election

		ct.uscongress.special.1.1801
			- Special election

		ct.congress.special.1805
			- Special election

		ga.uscongress.NorthernDistrict.1791
			- For Northern district

		ga.specialuscongress1.1806

		ky.uscongress3.1812
			- For third district

		ky.uscongress5.special.1810
		ky.specialuscongress1.1816
		me.uscongress4.secondrunoff.1821

		me.uscongress.3.2.1821
			- 2nd ballot
			- 3rd district

		me.uscongress3.third.1823
			- 3rd ballot
			- 3rd district

		md.uscongress5.special.18thcongress.1823

		ma.uscongress.2.hampshire.ballot2.1793
			- 2nd ballot
			- 2nd district

		ma.uscongress.eastern.2.1798
			- 1st ballot
			- Eastern 2 district

		ny.uscongressspecial.1791
	*/

again:

	recGroup, err := self.groupReader.read()
	if err != nil {
		return nil, err
	}

	// get state
	stateStr := self.getVal(recGroup[0], "state")
	state, ok := states.ByName[stateStr]
	if !ok {
		goto again
	}

	// get year
	yearStr := self.getVal(recGroup[0], "date")
	year := self.parseYear(yearStr)
	if year == nil {
		goto again
	}

	// get unique districts
	districtSet := make(map[string]bool)
	for _, rec := range recGroup {
		district := self.getVal(rec, "district")
		districtSet[district] = true
	}
	if len(districtSet) > 2 {
		goto again
	}
	if _, ok := districtSet[""]; !ok {
		goto again
	}

	// get district
	district := 0
	for districtStr := range districtSet {
		if districtStr == "" {
			continue
		}
		tmp := self.parseDistrict(districtStr)
		if tmp == nil {
			goto again
		}
		district = *tmp
	}

	// add up votes
	totalVotes := 0
	for _, rec := range recGroup {
		if self.getVal(rec, "district") != "" {
			continue
		}

		voteVal := self.getVal(rec, "Vote")
		if valIsNull(voteVal) {
			/* we're missing some data for this election */
			goto again
		}
		vote, err := strconv.Atoi(voteVal)
		if err != nil {
			goto again
		}
		totalVotes += vote
	}

	return &turnoutRec{
		stateUsps: state.Usps,
		district:  district,
		year:      *year,
		numVotes:  totalVotes}, nil
}

func (self *turnoutReader) parseYear(s string) *int {
	dashIdx := strings.IndexRune(s, '-')
	if dashIdx != -1 {
		s = s[:dashIdx]
	}
	year, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &year
}

func (self *turnoutReader) parseDistrict(s string) *int {
	district := 0
	if len(s) > 0 {
		ok := false
		district, ok = gWordsToNumbers[strings.ToLower(s)]
		if !ok {
			return nil
		}
	}
	return &district
}

func (self *turnoutReader) getVal(rec []string, col string) string {
	idx, ok := self.groupReader.colNameToIdx[strings.ToLower(col)]
	if !ok {
		log.Fatalf("Invalid column name: %v", col)
	}
	if idx >= len(rec) {
		log.Fatalf("Invalid column idx: %v", idx)
	}
	return strings.TrimSpace(rec[idx])
}

func GetData(ctx context.Context) (*sql.DB, error) {
	var err error
	var db *sql.DB
	var tx *sql.Tx
	var sourceInst *housedb.SourceInst
	var inserter bulkInserter.Inserter
	var reader *turnoutReader
	cols := []string{
		"district_nbr",
		"state",
		"congress_nbr",
		"turnout",
	}

	// connect to DB
	db, err = housedb.Connect()
	if err != nil {
		goto done
	}
	tx, err = housedb.StartTx(ctx, db)
	if err != nil {
		goto done
	}

	// get source
	sourceInst, err = housedb.FetchLocalSourceIfChanged(
		ctx,
		gSourceName,
		OpenTuftsData(),
		tx,
	)
	if err != nil {
		goto done
	}

	if sourceInst == nil {
		log.Printf("No new data for %v\n", gSourceName)
		goto done
	}
	log.Printf("New data for %v\n", gSourceName)
	defer sourceInst.Data.Close()

	// clear existing data
	_, err = tx.ExecContext(ctx, "DELETE FROM district_turnout")
	if err != nil {
		err = errors.Wrap(err, "Failed to delete")
		goto done
	}

	// process turnout data
	inserter = bulkInserter.Make(ctx, tx, "district_turnout", cols)
	reader = newTurnoutReader(sourceInst.Data)
	for {
		var data *turnoutRec
		data, err = reader.read()
		if err == io.EOF {
			break
		} else if err != nil {
			goto done
		}

		// get congress
		congress := congresses.GetForYear(data.year)
		if congress == nil {
			continue
		}

		// add turnout
		values := []interface{}{
			data.district,
			data.stateUsps,
			congress.Number,
			data.numVotes,
		}
		if err = inserter.Insert(values); err != nil {
			goto done
		}
	}

	// finish up
	if err = inserter.Flush(); err != nil {
		goto done
	}

	// mark source as processed
	if err = sourceInst.MakeRecord(); err != nil {
		goto done
	}

	// commit DB transaction
	if err = tx.Commit(); err != nil {
		err = errors.Wrap(err, "Failed to commit")
		goto done
	}
	tx = nil

done:
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		if db != nil {
			db.Close()
		}
		return nil, err
	}
	return db, nil
}
