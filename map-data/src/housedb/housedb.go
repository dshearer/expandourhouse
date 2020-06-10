package housedb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"expandourhouse.com/mapdata/housedb/reps"
	"expandourhouse.com/mapdata/housedb/turnout"
	_ "github.com/mattn/go-sqlite3"
)

const gSchema = `
PRAGMA journal_mode=WAL; /* Big performance improvement */

CREATE TABLE IF NOT EXISTS source(
	name TEXT NOT NULL,
	url TEXT,
	etag TEXT,
	last_checked INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS representative_term(
	bioguide TEXT NOT NULL, /* Unique ID of person */
	district_nbr INTEGER, /* NULL == unknown or at-large */
	at_large BOOL, /* NULL == unknown */
    state VARCHAR(2) NOT NULL, /* Two-digit state FIPS code */
    congress_nbr INTEGER NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,

	UNIQUE (bioguide, district_nbr, at_large, state, start_date, end_date),
	CONSTRAINT rep_term_dates CHECK (start_date <= end_date),
	CONSTRAINT district_nbr_is_pos CHECK (district_nbr > 0),
	CONSTRAINT district_nbr_at_large CHECK ((district_nbr IS NOT NULL) == (at_large IS FALSE))
);

CREATE TABLE IF NOT EXISTS district_turnout(
	district_nbr INTEGER NOT NULL, /* 0 == at-large */
    state VARCHAR(2) NOT NULL, /* Two-digit state FIPS code */
	congress_nbr INTEGER NOT NULL,
	turnout INTEGER NOT NULL,
	
	CONSTRAINT district_nbr_is_nonneg CHECK (district_nbr >= 0),
	CONSTRAINT turnout_is_nonneg CHECK (district_nbr >= 0)
);

CREATE VIEW IF NOT EXISTS state_with_overlapping_terms AS
SELECT DISTINCT t1.state, t1.congress_nbr
/* Collect all pairs of rep terms for the same state, district, and congress */
FROM representative_term t1 JOIN representative_term t2
	ON (t1.state = t2.state 
		AND t1.district_nbr = t2.district_nbr 
		AND t1.congress_nbr = t2.congress_nbr
		AND t1.bioguide != t2.bioguide
        AND t1.district_nbr IS NOT NULL)
/* Find ones that overlap in time */
WHERE t1.start_date <= t2.start_date AND t1.end_date > t2.start_date;

CREATE VIEW IF NOT EXISTS state_with_unknown_district AS
SELECT DISTINCT state, congress_nbr
FROM representative_term
WHERE district_nbr IS NULL;

CREATE VIEW IF NOT EXISTS state_with_atlarge_district AS
SELECT DISTINCT state, congress_nbr
FROM representative_term
WHERE at_large IS TRUE;

CREATE VIEW IF NOT EXISTS state_with_nonatlarge_district AS
SELECT DISTINCT state, congress_nbr
FROM representative_term
WHERE at_large IS FALSE;

CREATE VIEW IF NOT EXISTS state_with_atlarge_and_nonatlarge_districts AS
SELECT state, congress_nbr FROM state_with_atlarge_district
INTERSECT
SELECT state, congress_nbr FROM state_with_nonatlarge_district;

CREATE VIEW IF NOT EXISTS irregular_state AS
SELECT state, congress_nbr FROM state_with_atlarge_and_nonatlarge_districts
UNION
SELECT state, congress_nbr FROM state_with_unknown_district
UNION
SELECT state, congress_nbr FROM state_with_overlapping_terms;
`

const gDbPath = "./db"

type loadDataFunc func(context.Context, *sql.DB) error

var gLoadDataFuncs = []loadDataFunc{
	turnout.AddTurnoutData,
	reps.AddRepData,
}

func errIsDbLocked(err error) bool {
	return err != nil && strings.Index(err.Error(), "database is locked") != -1
}

type Db struct {
	db *sql.DB
	tx *sql.Tx
}

func Connect(ctx context.Context) Db {
	// connect to Db
	db, err := sql.Open("sqlite3", gDbPath)
	if err != nil {
		panic(err)
	}

	// make tables
	_, err = db.Exec(gSchema)
	if err != nil {
		panic(err)
	}

	// load data
	for _, f := range gLoadDataFuncs {
	again:
		if err := f(ctx, db); err != nil {
			if errIsDbLocked(err) {
				time.Sleep(3 * time.Second)
				goto again
			}
			panic(err)
		}
	}

	// make transaction for reading
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	return Db{db: db, tx: tx}
}

func (self *Db) Close() {
	if self.db == nil {
		return
	}
	self.tx.Rollback()
	self.db.Close()
	self.db = nil
	self.tx = nil
}

func (self *Db) NbrReps(ctx context.Context, congress int) int {
	res, err := self.db.Query(
		`SELECT COUNT(*) FROM representative_term WHERE congress = ? AND
		start_date = (SELECT MIN(start_date) FROM representative_term WHERE congress = ?)`,
		congress, congress)
	if err != nil {
		panic(err)
	}
	if !res.Next() {
		panic("Invalid congress nbr")
	}
	var nbr int
	if err := res.Scan(&nbr); err != nil {
		panic(err)
	}
	res.Close()
	return nbr
}

type irregType struct {
	name  string
	table string
}

var gIrregTypes = []irregType{
	irregType{
		name:  "Both at-large and on-at-large districts",
		table: "state_with_atlarge_and_nonatlarge_districts",
	},
	irregType{
		name:  "Districts have multiple reps",
		table: "state_with_overlapping_terms",
	},
}

func stateIsIrregular(db *sql.DB, state string, congressNbr int, it irregType) (bool, error) {
	sql := fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE state = $1 AND congress_nbr = $2", it.table)
	rows, err := db.Query(sql, state, congressNbr)
	if err != nil {
		return false, err
	}
	rows.Next()
	defer rows.Close()
	var nbr int
	if err := rows.Scan(&nbr); err != nil {
		return false, err
	}
	return nbr > 0, nil
}

func (self *Db) StateIrregularities(ctx context.Context, state string, congress int) []string {
	var irregHow []string
	for _, ir := range gIrregTypes {
		irreg, err := stateIsIrregular(
			self.db,
			state,
			congress,
			ir)
		if err != nil {
			panic(err)
		}
		if irreg {
			irregHow = append(irregHow, ir.name)
		}
	}
	return irregHow
}

// GetTurnout returns the turnout for the given district, if known.
// If the turnout is not known, it returns nil.
func (self *Db) GetTurnout(ctx context.Context, congress int, state string,
	district int) *int {

	if self.db == nil {
		panic("DB is closed")
	}

	query := "SELECT turnout FROM district_turnout WHERE congress_nbr = ? " +
		"AND state = ? AND district_nbr = ?"
	res, err := self.tx.QueryContext(ctx, query, congress, state, district)
	if err != nil {
		panic(err)
	}
	defer res.Close()
	if !res.Next() {
		return nil
	}
	var turnout int
	if err := res.Scan(&turnout); err != nil {
		panic(err)
	}
	return &turnout
}
