package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

var gDistrictCache = make(map[string]int)

func districtCacheKey(state string, district, year int) string {
	return fmt.Sprintf("%v%2d%4d", state, district, year)
}

func findDistrict(ctx context.Context, db *sql.DB, state string,
	district int, year int) (int, error) {

	var err error
	var districtId int
	var rows *sql.Rows
	var congressNbr int
	var sql string

	// check cache
	cacheKey := districtCacheKey(state, district, year)
	districtId, gotFromCache := gDistrictCache[cacheKey]
	if gotFromCache {
		goto done
	}

	// figure out which Congress we're talking about
	sql = "SELECT nbr FROM congress WHERE start_year > $1 ORDER BY start_year ASC"
	rows, err = db.QueryContext(ctx, sql, year)
	if err != nil {
		goto done
	}
	if !rows.Next() {
		err = fmt.Errorf("Couldn't find Congress for year %v", year)
		goto done
	}
	if err = rows.Scan(&congressNbr); err != nil {
		goto done
	}

	// find district
	rows.Close()
	sql = "SELECT id FROM house_district WHERE state = $1 AND district = $2 AND congress_nbr = $3"
	rows, err = db.QueryContext(ctx, sql, state, district, congressNbr)
	if err != nil {
		goto done
	}
	if rows.Next() {
		/* District exists */
		err = rows.Scan(&districtId)
		goto done
	}
	/* District doesn't exist */

	// make district
	rows.Close()
	sql = "INSERT INTO house_district(state, district, congress_nbr) VALUES ($1, $2, $3)" +
		" RETURNING id"
	rows, err = db.QueryContext(ctx, sql, state, district, congressNbr)
	if err != nil {
		goto done
	}
	rows.Next()
	err = rows.Scan(&districtId)

done:
	if rows != nil {
		rows.Close()
	}
	if !gotFromCache && err == nil {
		gDistrictCache[cacheKey] = districtId
	}
	return districtId, err
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

var gDistrictHasTurnoutCache = make(map[string]bool)

func districtHasTurnoutCacheKey(districtId, year int) string {
	return fmt.Sprintf("%5d%4d", districtId, year)
}

func districtHasTurnout(ctx context.Context, db *sql.DB, districtId,
	year int) (bool, error) {

	var hasTurnout bool
	var rows *sql.Rows
	var err error
	var sql string

	// check cache
	cacheKey := districtHasTurnoutCacheKey(districtId, year)
	hasTurnout, gotFromCache := gDistrictHasTurnoutCache[cacheKey]
	if gotFromCache {
		goto done
	}

	// check DB
	sql = "SELECT COUNT(*) > 0 FROM house_district_turnout " +
		"WHERE house_district_id = $1 AND year = $2"
	rows, err = db.QueryContext(ctx, sql, districtId, year)
	if err != nil {
		goto done
	}
	rows.Next()
	err = rows.Scan(&hasTurnout)

done:
	if rows != nil {
		rows.Close()
	}
	if !gotFromCache && err == nil {
		gDistrictHasTurnoutCache[cacheKey] = hasTurnout
	}
	return hasTurnout, err
}

// UpdateTurnout processes the turnout data
func UpdateTurnout(ctx context.Context, db *sql.DB, dataDirPath string) error {

	path := filepath.Join(dataDirPath, "house-election-results.csv")
	sourceText := "MIT Election Data and Science Lab, 2017, \"U.S. House 1976–2018\", " +
		"https://doi.org/10.7910/DVN/IG0UN2, Harvard Dataverse, V5, UNF:6:f4KhIVuYz/VinGbLYysWJg=="
	sourceId, err := getSource(ctx, db, sourceText)
	if err != nil {
		return err
	}

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
		districtHasTurnout, err := districtHasTurnout(ctx, db, districtId, data.year)
		if err != nil {
			return err
		}
		if districtHasTurnout {
			continue
		}

		// update DB
		sql := "INSERT INTO house_district_turnout(house_district_id, year, num_votes, source_id) " +
			"VALUES ($1, $2, $3, $4)"
		_, err = db.ExecContext(ctx, sql, districtId, data.year, data.totalVotes, sourceId)
		if err != nil {
			return err
		}

		// update cache
		cacheKey := districtHasTurnoutCacheKey(districtId, data.year)
		gDistrictHasTurnoutCache[cacheKey] = true

		n++
	} // for

	log.Printf("Inserted %v turnout records", n)
	return nil
}
