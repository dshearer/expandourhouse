package turnout

import (
	"context"
	"database/sql"
	"log"

	"expandourhouse.com/mapdata/bulkInserter"
	"expandourhouse.com/mapdata/congresses"
	"expandourhouse.com/mapdata/housedb/sourceinst"
)

type recKey struct {
	district int
	state    string
	congress int
}

func addHarvardData(ctx context.Context, tx *sql.Tx, source *sourceinst.SourceInst) error {
	// make reader
	reader := newTurnoutReader(source, ',')

	// make bulk-inserter
	inserter := bulkInserter.Make(ctx, tx, gTableName, gTableCols[:])

	seenRecs := make(map[recKey]bool)

	// read data
	for {
		// read record
		rec := reader.Read()
		if rec == nil {
			break
		}

		// skip non-House electionss
		if *rec.Get("office") != "US House" {
			continue
		}

		// skip special elections
		if rec.Get("stage") == nil ||
			*rec.Get("stage") != "gen" ||
			rec.Get("runoff") == nil ||
			*rec.Get("runoff") != "FALSE" ||
			rec.Get("special") == nil ||
			*rec.Get("special") != "FALSE" {
			continue
		}

		// get congress
		/*
			The election was either the year before the session of Congress,
			or in the same year.
		*/
		year := *rec.GetInt("year")
		if year%2 == 0 {
			year++
		}
		congress := congresses.GetForYear(year)
		if congress == nil {
			log.Printf("Can't find congress for year %v\n", year)
			continue
		}

		key := recKey{
			district: *rec.GetInt("district"),
			state:    *rec.Get("state_po"),
			congress: congress.Number,
		}
		if seenRecs[key] {
			continue
		}
		seenRecs[key] = true

		// insert total votes into DB
		values := []interface{}{
			*rec.GetInt("district"),
			*rec.Get("state_po"),
			congress.Number,
			*rec.GetInt("totalvotes"),
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
