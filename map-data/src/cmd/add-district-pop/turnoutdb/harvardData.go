package turnoutdb

import (
	"context"
	"database/sql"
	"log"

	"expandourhouse.com/mapdata/bulkInserter"
	"expandourhouse.com/mapdata/congresses"
	"expandourhouse.com/mapdata/housedb"
)

func addHarvardData(ctx context.Context, tx *sql.Tx, source *housedb.SourceInst) error {
	// make reader
	reader := newTurnoutReader(source, ',')

	// make bulk-inserter
	inserter := bulkInserter.Make(ctx, tx, gTableName, gTableCols[:])

	// read data
	for {
		// read record
		rec := reader.Read()
		if rec == nil {
			break
		}

		// skip non-House elections
		if rec.Get("office") != "US House" {
			continue
		}

		// skip special elections
		if rec.Get("stage") != "gen" || rec.Get("runoff") != "FALSE" ||
			rec.Get("special") != "FALSE" {
			continue
		}

		// get congress
		/*
			The election was either the year before the session of Congress,
			or in the same year.
		*/
		year := rec.GetInt("year")
		if year%2 == 0 {
			year++
		}
		congress := congresses.GetForYear(year)
		if congress == nil {
			log.Printf("Can't find congress for year %v\n", year)
			continue
		}

		// insert total votes into housedb
		values := []interface{}{
			rec.GetInt("district"),
			rec.Get("state_po"),
			congress.Number,
			rec.GetInt("totalvotes"),
		}
		if err := inserter.Insert(values); err != nil {
			return err
		}
	} // for

	// finish up
	if err := inserter.Flush(); err != nil {
		return err
	}

	return nil
}
