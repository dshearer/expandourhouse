package tuftsTurnout

import (
	"context"
	"database/sql"
	"encoding/csv"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"expandourhouse.com/loaddata/utils"
)

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

func newTurnoutReader(f *os.File) *turnoutReader {
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

func ProcessTuftsTurnout(ctx context.Context, db *sql.DB, dataDirPath string) error {
	dataPath := path.Join(dataDirPath, "tufts-all-votes-congress-3.tsv")
	log.Printf("Processing %v", dataPath)

	// load irregular states
	if err := loadIrregStates(ctx, db); err != nil {
		return err
	}

	// make source row
	sourceText := "Lampi Collection of American Electoral Returns, 1787–1825. American Antiquarian Society, 2007."
	sourceId, err := utils.GetSource(ctx, db, sourceText)
	if err != nil {
		return err
	}

	// open data file
	f, err := os.Open(dataPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// read it
	var n int
	reader := newTurnoutReader(f)
	for {
		data, err := reader.read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// get congress nbr
		if data.year%2 == 0 {
			data.year++
		}
		congressNbr, err := utils.GetCongressNbr(ctx, db, data.year)
		if err != nil {
			return err
		}

		// check if this state is irregular
		if isIrregState(data.stateUsps, congressNbr) {
			continue
		}

		// find district
		districtId, err := utils.GetDistrict(ctx, db, data.stateUsps, data.district, congressNbr)
		if err != nil {
			return err
		}

		// check if we already have data for this district
		districtHasTurnout, err := utils.DistrictHasTurnout(ctx, db, districtId, sourceId)
		if err != nil {
			return err
		}
		if districtHasTurnout {
			continue
		}

		// update DB
		err = utils.AddDistrictTurnout(ctx, db, districtId, sourceId, data.numVotes)
		if err != nil {
			return err
		}

		n++
	}

	log.Printf("Inserted %v turnout records", n)
	return nil
}
