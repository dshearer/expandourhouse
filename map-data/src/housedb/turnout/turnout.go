//go:generate go run scripts/includeData.go

package turnout

import (
	"context"
	"database/sql"
	"log"

	"expandourhouse.com/mapdata/bulkInserter"
	"expandourhouse.com/mapdata/housedb/sourceinst"
)

const gTuftsSourceName = "tufts-turnout"
const gHarvardSourceName = "harvard-turnout"

var gTableCols = [...]string{
	"district_nbr",
	"state",
	"congress_nbr",
	"turnout",
}

const gTuftsTurnoutTable = "tufts_district_turnout"

func addTuftsData(ctx context.Context, tx *sql.Tx, source *sourceinst.SourceInst) error {
	// delete old data
	if _, err := tx.ExecContext(ctx, "DELETE FROM "+gTuftsTurnoutTable); err != nil {
		return err
	}

	inserter := bulkInserter.Make(ctx, tx, gTuftsTurnoutTable, gTableCols[:])
	reader := newTurnoutReader(source, ',')
	for {
		rec := reader.Read()
		if rec == nil {
			break
		}
		values := []interface{}{
			*rec.GetInt("district"),
			*rec.Get("state"),
			*rec.GetInt("congress_nbr"),
			*rec.GetInt("vote"),
		}
		if err := inserter.Insert(values); err != nil {
			return err
		}
	}
	return inserter.Flush()
}

func AddTurnoutData(ctx context.Context, db *sql.DB) error {
	var tx *sql.Tx
	var err error

	defer func() {
		if err != nil {
			if tx != nil {
				tx.Rollback()
			}
		}
	}()

	// make transaction
	tx, err = db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	if err != nil {
		return err
	}

	// process Tufts data
	var sourceInst *sourceinst.SourceInst
	sourceInst, err = sourceinst.FetchLocalSourceIfChanged(
		ctx,
		gTuftsSourceName,
		OpenTuftsData(),
		tx,
	)
	if err != nil {
		return err
	}
	if sourceInst == nil {
		log.Printf("No new data for %v\n", gTuftsSourceName)
	} else {
		// add data to DB
		defer sourceInst.Data.Close()
		log.Printf("New data for %v\n", gTuftsSourceName)
		if err = addTuftsData(ctx, tx, sourceInst); err != nil {
			return err
		}

		// mark source as processed
		if err = sourceInst.MakeRecord(); err != nil {
			return err
		}
	}

	// commit DB transaction
	if err = tx.Commit(); err != nil {
		return err
	}

	// make transaction
	tx, err = db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	if err != nil {
		return err
	}

	// process Havard data
	sourceInst, err = sourceinst.FetchLocalSourceIfChanged(
		ctx,
		gHarvardSourceName,
		OpenHarvardData(),
		tx,
	)
	if err != nil {
		return err
	}
	if sourceInst == nil {
		log.Printf("No new data for %v\n", gHarvardSourceName)
	} else {
		// add data to DB
		defer sourceInst.Data.Close()
		log.Printf("New data for %v\n", gHarvardSourceName)
		if err = addHarvardData(ctx, tx, sourceInst); err != nil {
			return err
		}

		// mark source as processed
		if err = sourceInst.MakeRecord(); err != nil {
			return err
		}
	}

	// commit DB transaction
	if err = tx.Commit(); err != nil {
		return err
	}

	return err
}
