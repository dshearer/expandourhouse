package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"expandourhouse.com/mapdata/bulkInserter"
	"expandourhouse.com/mapdata/congresses"
	districtdb "expandourhouse.com/mapdata/districtDb"
	"gopkg.in/yaml.v2"
)

const gSourceName = "legislators-historical"
const gSourceUrl = "https://github.com/unitedstates/congress-legislators/raw/master/legislators-historical.yaml"

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

func handleHistLegEntry(ctx context.Context, db *sql.DB,
	entry map[interface{}]interface{},
	inserter *bulkInserter.Inserter) error {

	id := entry["id"].(map[interface{}]interface{})
	terms := entry["terms"].([]interface{})
	for _, e := range terms {
		term := e.(map[interface{}]interface{})

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

		/*
			district == -1 --> Indian territory
			district == 0 --> at-large
		*/

		bioguide := id["bioguide"].(string)
		congress := congresses.GetForYear(start.Year())
		state := term["state"].(string)
		districtNbr := term["district"].(int)
		var districtNbrP *int
		var atLargeP *bool
		if districtNbr > 0 {
			districtNbrP = &districtNbr
			tmp := false
			atLargeP = &tmp
		} else if districtNbr == 0 {
			/* At-large district */
			tmp := true
			atLargeP = &tmp
		} else {
			/* Indian territory */
			continue
		}

		// insert into DB
		values := []interface{}{
			bioguide,
			districtNbrP,
			atLargeP,
			state,
			congress.Number,
			start,
			end,
		}
		if err = inserter.Insert(values); err != nil {
			return err
		}
	}

	return nil
}

func GetData(ctx context.Context) (*sql.DB, error) {
	// connect to DB
	db, err := districtdb.Connect()
	if err != nil {
		return nil, err
	}

	// get source
	sourceInst, err := FetchSourceIfChanged(
		ctx,
		gSourceName,
		gSourceUrl,
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

	// parse as YAML
	decoder := yaml.NewDecoder(sourceInst.Data)
	var data []map[interface{}]interface{}
	err = decoder.Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	// empty DB
	_, err = db.ExecContext(ctx, "DELETE FROM representative_term")
	if err != nil {
		log.Fatal(err)
	}

	// add entries to DB
	log.Println("Adding historial legislators")
	cols := []string{
		"bioguide",
		"district_nbr",
		"at_large",
		"state",
		"congress_nbr",
		"start_date",
		"end_date",
	}
	inserter := bulkInserter.Make(ctx, db, "representative_term", cols)
	for _, entry := range data {
		err = handleHistLegEntry(ctx, db, entry, &inserter)
		if err != nil {
			log.Fatal(err)
		}
	}
	if err = inserter.Flush(); err != nil {
		log.Fatal(err)
	}

	// mark source as processed
	if err := sourceInst.MakeRecord(); err != nil {
		return nil, err
	}

	return db, nil
}
