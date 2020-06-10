//go:generate go run scripts/includeData.go

package turnout

import (
	"context"
	"database/sql"
	"log"

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

const gTableName = "district_turnout"

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
	tx, err = db.BeginTx(ctx, nil)
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
