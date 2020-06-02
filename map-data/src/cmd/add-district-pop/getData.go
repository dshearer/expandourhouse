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
)

const gSourceName = "tufts-turnout"

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

type turnoutReader struct {
	csvReader    *csv.Reader
	colNameToIdx map[string]int

	// record context:
	currId         *string
	currTurnoutRec turnoutRec
	recBuff        []string
}

func newTurnoutReader(f io.Reader) *turnoutReader {
	r := &turnoutReader{csv.NewReader(f), nil, nil, turnoutRec{}, nil}
	r.csvReader.ReuseRecord = true
	r.csvReader.Comma = '\t'
	return r
}

func (self *turnoutReader) read() (*turnoutRec, error) {
	for {
		var action *readerAction
		var err error
		if self.currId == nil {
			action, err = self.read_noCurrId()
		} else {
			action, err = self.read_currId()
		}
		if err != nil {
			return nil, err
		}

		switch *action {
		case gReaderActionReturnRec:
			self.currId = nil
			return &self.currTurnoutRec, nil

		case gReaderActionAbandonRec:
			self.currId = nil

		case gReaderActionAgain:
			continue
		} // switch
	} // for
}

func (self *turnoutReader) getVal(rec []string, col string) string {
	idx, ok := self.colNameToIdx[col]
	if !ok {
		log.Fatalf("Invalid column name: %v", col)
	}
	if idx >= len(rec) {
		log.Fatalf("Invalid column idx: %v", idx)
	}
	return rec[idx]
}

func (self *turnoutReader) readNextLine() ([]string, error) {
	if self.recBuff != nil {
		rec := self.recBuff
		self.recBuff = nil
		return rec, nil
	}

	return self.csvReader.Read()
}

func (self *turnoutReader) read_noCurrId() (*readerAction, error) {
	// get next line
	rec, err := self.readNextLine()
	if err != nil {
		return nil, err
	}
	if len(rec) == 0 {
		action := gReaderActionAgain
		return &action, nil
	}

	if self.colNameToIdx == nil {
		// keep column names
		self.colNameToIdx = make(map[string]int)
		for idx, colName := range rec {
			self.colNameToIdx[strings.TrimSpace(colName)] = idx
		}
		action := gReaderActionAgain
		return &action, nil
	}

	// parse id
	id := self.getVal(rec, "id")
	parts := strings.Split(id, ".")
	if len(parts) != 4 || parts[1] != "uscongress" {
		/* We don't care about this row */
		action := gReaderActionAgain
		return &action, nil
	}
	stateUsps := strings.ToUpper(parts[0])
	district, err1 := strconv.Atoi(parts[2])
	year, err2 := strconv.Atoi(parts[3])
	if err1 != nil || err2 != nil {
		/* We don't care about this row */
		action := gReaderActionAgain
		return &action, nil
	}

	// begin making a record
	self.currId = &id
	self.currTurnoutRec.init(stateUsps, district, year)

	// push current rec into buffer so that we will process its vote count
	self.recBuff = rec

	action := gReaderActionAgain
	return &action, nil
}

func (self *turnoutReader) read_currId() (*readerAction, error) {
	// get next line
	rec, err := self.readNextLine()
	if err != nil {
		return nil, err
	}
	if len(rec) == 0 {
		action := gReaderActionAgain
		return &action, nil
	}

	// check id
	if self.getVal(rec, "id") != *self.currId {
		/* Done with this record. */

		// push rec into buffer so we'll process it later
		self.recBuff = rec

		// return action
		var action readerAction
		if self.currTurnoutRec.numVotes == 0 {
			action = gReaderActionAbandonRec
		} else {
			action = gReaderActionReturnRec
		}
		return &action, nil
	}

	// check place names
	for _, placeCol := range gPlaceCols {
		val := self.getVal(rec, placeCol)
		if !valIsNull(val) {
			/* Done with this record. */
			var action readerAction
			if self.currTurnoutRec.numVotes == 0 {
				action = gReaderActionAbandonRec
			} else {
				action = gReaderActionReturnRec
			}
			return &action, nil
		}
	} // for

	// add to vote count
	voteVal := self.getVal(rec, "Vote")
	if valIsNull(voteVal) {
		/* we're missing some data for this election */
		action := gReaderActionAbandonRec
		return &action, nil
	}
	vote, err := strconv.Atoi(voteVal)
	if err != nil {
		return nil, err
	}
	self.currTurnoutRec.numVotes += vote
	action := gReaderActionAgain
	return &action, nil
}

func GetData(ctx context.Context) (*sql.DB, error) {
	// connect to DB
	db, err := housedb.Connect()
	if err != nil {
		return nil, err
	}

	// get source
	sourceInst, err := housedb.FetchLocalSourceIfChanged(
		ctx,
		gSourceName,
		OpenTuftsData(),
		db,
	)
	if err != nil {
		return nil, err
	}

	if sourceInst == nil {
		log.Printf("No new data for %v\n", gSourceName)
		return db, nil
	}
	log.Printf("New data for %v\n", gSourceName)
	defer sourceInst.Data.Close()

	// clear existing data
	_, err = db.ExecContext(ctx, "DELETE FROM district_turnout")
	if err != nil {
		return nil, err
	}

	// process turnout data
	cols := []string{
		"district_nbr",
		"state",
		"congress_nbr",
		"turnout",
	}
	inserter := bulkInserter.Make(ctx, db, "district_turnout", cols)
	reader := newTurnoutReader(sourceInst.Data)
	for {
		data, err := reader.read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
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
			return nil, err
		}
	}

	// finish up
	if err = inserter.Flush(); err != nil {
		return nil, err
	}

	// mark source as processed
	if err := sourceInst.MakeRecord(); err != nil {
		return nil, err
	}

	return db, nil
}
