package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"expandourhouse.com/loaddata/bulkInserter"
)

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

func findCongress(ctx context.Context, db *sql.DB, date time.Time) (*int, error) {
	var nbr *int
	var rows *sql.Rows
	var err error
	var tmp int

	year := date.Year()
	if year%2 == 0 {
		year--
	}

	sql := "SELECT nbr FROM congress WHERE start_year = $1"
	rows, err = db.QueryContext(ctx, sql, year)
	if err != nil {
		goto done
	}
	if !rows.Next() {
		goto done
	}
	err = rows.Scan(&tmp)
	if err != nil {
		goto done
	}
	nbr = &tmp

done:
	if rows != nil {
		rows.Close()
	}
	return nbr, err
}

func handleHistLegEntry(ctx context.Context, db *sql.DB,
	entry map[string]interface{}, inserter *bulkInserter.Inserter) error {

	id := entry["id"].(map[string]interface{})
	bioguide := id["bioguide"].(string)

	terms := entry["terms"].([]interface{})
	for _, e := range terms {
		term := e.(map[string]interface{})

		// check type
		typ := term["type"].(string)
		if typ != "rep" {
			continue
		}

		// parse dates
		start, err := parseDate(term["start"].(string))
		if err != nil {
			return err
		}
		end, err := parseDate(term["end"].(string))
		if err != nil {
			return err
		}

		// find congress
		congressNbr, err := findCongress(ctx, db, start)
		if err != nil {
			return err
		}
		if congressNbr == nil {
			log.Printf("Failed to find congress for year %v", start.Year())
			continue
		}

		// get state
		state := term["state"].(string)

		// find district
		districtNbr := int(term["district"].(float64))
		var districtId *int
		if districtNbr != -1 {
			tmp, err := getDistrict(ctx, db, state, districtNbr, *congressNbr)
			if err != nil {
				return err
			}
			districtId = &tmp
		}

		// insert into DB
		values := []interface{}{districtId, start, end, bioguide, state,
			congressNbr}
		if err = inserter.Insert(values); err != nil {
			return err
		}
	}

	return nil
}

func UpdateHistoricalLegislators(ctx context.Context, db *sql.DB, dataDirPath string) error {
	// parse JSON
	statesFilePath := filepath.Join(dataDirPath, "legislators-historical.json")
	f, err := os.Open(statesFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	var data []map[string]interface{}
	err = decoder.Decode(&data)
	if err != nil {
		return err
	}

	// empty DB
	_, err = db.ExecContext(ctx, "TRUNCATE representative_term")
	if err != nil {
		return err
	}

	// add entries to DB
	log.Print("Adding historial legislators")
	cols := []string{"house_district_id", "start_date", "end_date",
		"bioguide_id", "state", "congress_nbr"}
	inserter := bulkInserter.Make(ctx, db, "representative_term", cols)
	for _, entry := range data {
		err = handleHistLegEntry(ctx, db, entry, &inserter)
		if err != nil {
			return err
		}
	}
	if err = inserter.Flush(); err != nil {
		return err
	}

	// refresh materialized view that finds the "irregular" states
	_, err = db.ExecContext(ctx, "REFRESH MATERIALIZED VIEW irregular_state")
	if err != nil {
		return err
	}

	return nil
}
