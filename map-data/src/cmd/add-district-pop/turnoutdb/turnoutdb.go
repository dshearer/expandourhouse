//go:generate go run ../scripts/includeData.go

package turnoutdb

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"expandourhouse.com/mapdata/housedb"
	"github.com/pkg/errors"
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

var ErrTryAgain = errors.New("Try again")

type TurnoutDb struct {
	db *sql.DB
	tx *sql.Tx
}

func NewTurnoutDb(ctx context.Context) (*TurnoutDb, error) {
	var db *sql.DB
	var tx *sql.Tx
	var err error

	defer func() {
		if err != nil {
			if tx != nil {
				tx.Rollback()
			}
			if db != nil {
				db.Close()
			}
		}
	}()

	// connect to DB
	db, err = housedb.Connect()
	if err != nil {
		panic(err)
	}
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	// process Tufts data
	var sourceInst *housedb.SourceInst
	sourceInst, err = housedb.FetchLocalSourceIfChanged(
		ctx,
		gTuftsSourceName,
		OpenTuftsData(),
		tx,
	)
	if err != nil {
		panic(err)
	}
	if sourceInst == nil {
		log.Printf("No new data for %v\n", gTuftsSourceName)
	} else {
		// add data to DB
		defer sourceInst.Data.Close()
		log.Printf("New data for %v\n", gTuftsSourceName)
		if err = addTuftsData(ctx, tx, sourceInst); err != nil {
			if housedb.ErrIsDbLocked(err) {
				err = ErrTryAgain
				return nil, err
			}
			panic(err)
		}

		// mark source as processed
		if err = sourceInst.MakeRecord(); err != nil {
			if housedb.ErrIsDbLocked(err) {
				err = ErrTryAgain
				return nil, err
			}
			panic(err)
		}
	}

	// process Havard data
	sourceInst, err = housedb.FetchLocalSourceIfChanged(
		ctx,
		gHarvardSourceName,
		OpenHarvardData(),
		tx,
	)
	if err != nil {
		panic(err)
	}
	if sourceInst == nil {
		log.Printf("No new data for %v\n", gHarvardSourceName)
	} else {
		// add data to DB
		defer sourceInst.Data.Close()
		log.Printf("New data for %v\n", gHarvardSourceName)
		if err = addHarvardData(ctx, tx, sourceInst); err != nil {
			if housedb.ErrIsDbLocked(err) {
				err = ErrTryAgain
				return nil, err
			}
			panic(err)
		}

		// mark source as processed
		if err = sourceInst.MakeRecord(); err != nil {
			if housedb.ErrIsDbLocked(err) {
				err = ErrTryAgain
				return nil, err
			}
			panic(err)
		}
	}

	// commit DB transaction
	if err = tx.Commit(); err != nil {
		if housedb.ErrIsDbLocked(err) {
			err = ErrTryAgain
			return nil, err
		}
		panic(err)
	}
	tx = nil

	// make transaction just for reading
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	return &TurnoutDb{db: db, tx: tx}, nil
}

// GetTurnout returns the turnout for the given district, if known.
// If the turnout is not known, it returns (nil, nil).
func (self *TurnoutDb) GetTurnout(ctx context.Context, congress int,
	state string, district int) (*int, error) {

	if self.db == nil {
		return nil, fmt.Errorf("TurnoutDB is closed")
	}

	query := "SELECT turnout FROM district_turnout WHERE congress_nbr = ? " +
		"AND state = ? AND district_nbr = ?"
	res, err := self.tx.QueryContext(ctx, query, congress, state, district)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	if !res.Next() {
		return nil, nil
	}
	var turnout int
	if err := res.Scan(&turnout); err != nil {
		return nil, err
	}
	return &turnout, nil
}

func (self *TurnoutDb) Close() {
	if self.db == nil {
		return
	}
	self.tx.Rollback()
	self.db.Close()
	self.db = nil
	self.tx = nil
}
