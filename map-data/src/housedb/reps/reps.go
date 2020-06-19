package reps

import (
	"context"
	"database/sql"
	"log"
	"time"

	"expandourhouse.com/mapdata/bulkInserter"
	"expandourhouse.com/mapdata/congresses"
	"expandourhouse.com/mapdata/housedb/sourceinst"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var gRepSources = [][]string{
	{"legislators-historical", "https://github.com/unitedstates/congress-legislators/raw/master/legislators-historical.yaml"},
	{"legislators-current", "https://github.com/unitedstates/congress-legislators/raw/master/legislators-current.yaml"},
}

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

func handleHistLegEntry(ctx context.Context, db *sql.Tx,
	entry map[interface{}]interface{},
	inserter *bulkInserter.Inserter) error {

	id := entry["id"].(map[interface{}]interface{})
	name := entry["name"].(map[interface{}]interface{})
	firstName := name["first"].(string)
	middleName, hasMiddleName := name["middle"].(string)
	lastName := name["last"].(string)

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
		var middleNameP *string
		if hasMiddleName {
			middleNameP = &middleName
		}
		values := []interface{}{
			bioguide,
			firstName,
			middleNameP,
			lastName,
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

func addRepData(ctx context.Context, tx *sql.Tx, sourceInst *sourceinst.SourceInst) error {
	cols := []string{
		"bioguide",
		"first_name",
		"middle_name",
		"last_name",
		"district_nbr",
		"at_large",
		"state",
		"congress_nbr",
		"start_date",
		"end_date",
	}

	// parse as YAML
	decoder := yaml.NewDecoder(sourceInst.Data)
	var data []map[interface{}]interface{}
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	// add entries to DB
	inserter := bulkInserter.Make(ctx, tx, "representative_term", cols)
	inserter.FlushPeriod = 75
	for _, entry := range data {
		if err := handleHistLegEntry(ctx, tx, entry, &inserter); err != nil {
			return err
		}
	}
	if err := inserter.Flush(); err != nil {
		err = errors.Wrap(err, "Failed to flush")
		return err
	}

	// mark source as processed
	if err := sourceInst.MakeRecord(); err != nil {
		err = errors.Wrap(err, "Failed to make source record")
		return err
	}
	return nil
}

func AddRepData(ctx context.Context, db *sql.DB) (err error) {
	var tx *sql.Tx

	defer func() {
		if err != nil {
			if tx != nil {
				log.Println("reps rolling back")
				tx.Rollback()
			}
		}
	}()

	clearedTable := false

	for _, pair := range gRepSources {
		sourceName := pair[0]
		sourceUrl := pair[1]

		// start transaction
		tx, err = db.BeginTx(ctx, nil)
		if err != nil {
			return
		}

		// get source
		var sourceInst *sourceinst.SourceInst
		sourceInst, err = sourceinst.FetchHttpSourceIfChanged(
			ctx,
			sourceName,
			sourceUrl,
			tx,
		)
		if err != nil {
			tx.Rollback()
			return
		}

		if sourceInst == nil {
			log.Printf("No new data for %v\n", sourceName)
		} else {
			defer sourceInst.Data.Close()
			log.Printf("New data for %v\n", sourceName)

			if !clearedTable {
				if _, err = tx.ExecContext(ctx, "DELETE FROM representative_term"); err != nil {
					err = errors.Wrap(err, "Failed to delete")
					tx.Rollback()
					return
				}
				clearedTable = true
			}

			if err = addRepData(ctx, tx, sourceInst); err != nil {
				tx.Rollback()
				return
			}
		}

		// commit DB transaction
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return
		}
	}

	return
}
