package mitTurnout

import (
	"context"
	"database/sql"
	"encoding/csv"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"expandourhouse.com/loaddata/utils"
)

func findDistrict(ctx context.Context, db *sql.DB, state string,
	district int, year int) (int, error) {

	// figure out which Congress we're talking about
	congressNbr, err := utils.GetCongressNbr(ctx, db, year+1)
	if err != nil {
		return 0, err
	}

	// find district
	return utils.GetDistrict(ctx, db, state, district, congressNbr)
}

type turnoutData struct {
	year       int
	statePo    string
	district   int
	totalVotes int
}

type turnoutReader struct {
	csvReader    *csv.Reader
	colNameToIdx map[string]int
}

func newTurnoutReader(f *os.File) *turnoutReader {
	r := &turnoutReader{csv.NewReader(f), nil}
	r.csvReader.ReuseRecord = true
	return r
}

func (self *turnoutReader) read() (*turnoutData, error) {
	var rec []string
	var err error
	var data turnoutData

do:
	rec, err = self.csvReader.Read()
	if err != nil {
		return nil, err
	}
	if len(rec) == 0 {
		goto do
	}

	if self.colNameToIdx == nil {
		// keep column names
		self.colNameToIdx = make(map[string]int)
		for idx, colName := range rec {
			self.colNameToIdx[colName] = idx
		}
		goto do
	}

	yearIdx := self.colNameToIdx["year"]
	stateIdx := self.colNameToIdx["state_po"]
	districtIdx := self.colNameToIdx["district"]
	totalVotesIdx := self.colNameToIdx["totalvotes"]

	data.year, err = strconv.Atoi(rec[yearIdx])
	if err != nil {
		// return nil, errors.Wrap(err, "Problem with year value")
		goto do
	}
	data.statePo = rec[stateIdx]
	data.district, err = strconv.Atoi(rec[districtIdx])
	if err != nil {
		// return nil, errors.Wrap(err, "Problem with district value")
		goto do
	}
	data.totalVotes, err = strconv.Atoi(rec[totalVotesIdx])
	if err != nil || data.totalVotes < 2 {
		goto do
	}

	return &data, nil
}

// UpdateTurnout processes the turnout data
func UpdateTurnout(ctx context.Context, db *sql.DB, dataDirPath string) error {

	path := filepath.Join(dataDirPath, "house-election-results.csv")

	// make source row
	sourceText := "MIT Election Data and Science Lab, 2017, \"U.S. House 1976–2018\", " +
		"https://doi.org/10.7910/DVN/IG0UN2, Harvard Dataverse, V5, UNF:6:f4KhIVuYz/VinGbLYysWJg=="
	sourceId, err := utils.GetSource(ctx, db, sourceText)
	if err != nil {
		return err
	}

	// open data file
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	n := 0
	reader := newTurnoutReader(f)
	for {
		/*
			NOTE: The data file has multiple entries for one district,
			bc it has an entry for each candidate.
		*/

		data, err := reader.read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// find district
		districtId, err := findDistrict(ctx, db, data.statePo, data.district, data.year)
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
		err = utils.AddDistrictTurnout(ctx, db, districtId, sourceId, data.totalVotes)
		if err != nil {
			return err
		}

		n++
	} // for

	log.Printf("Inserted %v turnout records", n)
	return nil
}
