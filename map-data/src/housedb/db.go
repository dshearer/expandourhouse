package housedb

import (
	"context"
	"database/sql"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
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

func Connect() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", gDbPath)
	if err != nil {
		return nil, err
	}

	// make tables
	_, err = db.Exec(gSchema)
	if err != nil {
		db.Close()
		return nil, errors.Wrap(err, "Failed to make tables")
	}

	return db, nil
}

func ErrIsDbLocked(err error) bool {
	return err != nil && strings.Index(err.Error(), "database is locked") != -1
}

func StartTx(ctx context.Context, db *sql.DB) (*sql.Tx, error) {
	var tx *sql.Tx
	var err error
	stopTrying := time.Now().Add(5 * time.Minute)
	for time.Now().Before(stopTrying) {
		tx, err = db.BeginTx(ctx, nil)
		if err == nil {
			return tx, nil
		}
		time.Sleep(3 * time.Second)
	}
	return nil, err
}
