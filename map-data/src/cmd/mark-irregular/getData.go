package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"expandourhouse.com/mapdata/bulkInserter"
	"expandourhouse.com/mapdata/congresses"
	"expandourhouse.com/mapdata/housedb"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const gSourceName = "legislators-historical"
const gSourceUrl = "https://github.com/unitedstates/congress-legislators/raw/master/legislators-historical.yaml"

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

func handleHistLegEntry(ctx context.Context, db *sql.Tx,
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
	var err error
	var db *sql.DB
	var tx *sql.Tx
	var sourceInst *housedb.SourceInst
	var inserter bulkInserter.Inserter
	var data []map[interface{}]interface{}
	var decoder *yaml.Decoder
	cols := []string{
		"bioguide",
		"district_nbr",
		"at_large",
		"state",
		"congress_nbr",
		"start_date",
		"end_date",
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
	sourceInst, err = housedb.FetchHttpSourceIfChanged(
		ctx,
		gSourceName,
		gSourceUrl,
		tx,
	)
	if err != nil {
		goto done
	}

	if sourceInst == nil {
		log.Printf("No new data for %v\n", gSourceName)
		tx.Rollback()
		goto done
	}
	log.Printf("New data for %v\n", gSourceName)
	defer sourceInst.Data.Close()

	// parse as YAML
	decoder = yaml.NewDecoder(sourceInst.Data)
	err = decoder.Decode(&data)
	if err != nil {
		goto done
	}

	// empty DB
	_, err = tx.ExecContext(ctx, "DELETE FROM representative_term")
	if err != nil {
		err = errors.Wrap(err, "Failed to delete")
		goto done
	}

	// add entries to DB
	log.Println("Adding historial legislators")
	inserter = bulkInserter.Make(ctx, tx, "representative_term", cols)
	for _, entry := range data {
		err = handleHistLegEntry(ctx, tx, entry, &inserter)
		if err != nil {
			goto done
		}
	}
	if err = inserter.Flush(); err != nil {
		err = errors.Wrap(err, "Failed to flush")
		goto done
	}

	// mark source as processed
	if err = sourceInst.MakeRecord(); err != nil {
		err = errors.Wrap(err, "Failed to make source record")
		goto done
	}

	// commit DB transaction
	if err = tx.Commit(); err != nil {
		err = errors.Wrap(err, "Failed to commit")
		goto done
	}

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
