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
	name TEXT NOT NULL UNIQUE,
	url TEXT,
	etag TEXT,
	last_checked INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS representative_term(
	bioguide TEXT NOT NULL, /* Unique ID of person */
	first_name TEXT NOT NULL,
	middle_name TEXT,
	last_name TEXT NOT NULL,
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

CREATE TABLE IF NOT EXISTS tufts_district_turnout(
	district_nbr INTEGER NOT NULL, /* 0 == at-large */
    state VARCHAR(2) NOT NULL, /* Two-digit state FIPS code */
	congress_nbr INTEGER NOT NULL,
	turnout INTEGER NOT NULL,
	
	UNIQUE (district_nbr, state, congress_nbr),
	CONSTRAINT district_nbr_is_nonneg CHECK (district_nbr >= 0),
	CONSTRAINT turnout_is_nonneg CHECK (district_nbr >= 0)
);

CREATE TABLE IF NOT EXISTS harvard_district_turnout(
	district_nbr INTEGER NOT NULL, /* 0 == at-large */
    state VARCHAR(2) NOT NULL, /* Two-digit state FIPS code */
	congress_nbr INTEGER NOT NULL,
	turnout INTEGER NOT NULL,
	
	UNIQUE (district_nbr, state, congress_nbr),
	CONSTRAINT district_nbr_is_nonneg CHECK (district_nbr >= 0),
	CONSTRAINT turnout_is_nonneg CHECK (district_nbr >= 0)
);

CREATE VIEW IF NOT EXISTS district_turnout AS
SELECT * FROM tufts_district_turnout WHERE turnout > 10
UNION
SELECT * FROM harvard_district_turnout WHERE turnout > 10;

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
	db                      *sql.DB
	tx                      *sql.Tx
	medianVotersPerDistrict map[int]float64
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
		tries := 0
	again:
		if err := f(ctx, db); err != nil {
			if errIsDbLocked(err) {
				time.Sleep(3 * time.Second)
				tries++
				if tries > 10 {
					panic("Can't unlock DB")
				}
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
	if congress >= 61 {
		return 435
	}

	res, err := self.db.Query(
		`SELECT COUNT(*) FROM representative_term WHERE congress_nbr = ? AND
		start_date = (SELECT MIN(start_date) FROM representative_term WHERE congress_nbr = ?)`,
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

func median(values []int) float64 {
	mid := len(values) / 2
	if len(values)%2 == 0 {
		return (float64(values[mid-1]) + float64(values[mid])) / float64(2)
	} else {
		return float64(values[mid])
	}
}

func (self *Db) precomputeMedianVotersPerRegDistrict(ctx context.Context) {
	if self.medianVotersPerDistrict != nil {
		return
	}

	q := `
	SELECT dt.congress_nbr, dt.turnout
	FROM district_turnout dt LEFT OUTER JOIN irregular_state ir 
	ON (dt.state = ir.state AND dt.congress_nbr = ir.congress_nbr)
	WHERE ir.state IS NULL
	ORDER BY dt.congress_nbr, dt.turnout ASC
	`
	res, err := self.db.QueryContext(ctx, q)
	if err != nil {
		panic(err)
	}
	defer res.Close()

	// organize turnouts by congress
	turnoutMap := make(map[int][]int)
	for res.Next() {
		var congress int
		var turnout int
		if err := res.Scan(&congress, &turnout); err != nil {
			panic(err)
		}
		array, ok := turnoutMap[congress]
		if !ok {
			array = nil
			turnoutMap[congress] = array
		}
		turnoutMap[congress] = append(turnoutMap[congress], turnout)
	}

	// compute medians
	self.medianVotersPerDistrict = make(map[int]float64)
	for congress, turnout := range turnoutMap {
		self.medianVotersPerDistrict[congress] = median(turnout)
	}
}

func (self *Db) MedianVotersPerRegDistrict(ctx context.Context, congress int) (float64, bool) {
	self.precomputeMedianVotersPerRegDistrict(ctx)
	val, ok := self.medianVotersPerDistrict[congress]
	if !ok {
		return 0, false
	}
	return val, true
}

func (self *Db) _STATVotersPerRegDistrict(ctx context.Context, congress int, stat string) (float64, bool) {
	q := `
	SELECT %s(dt.turnout)
	FROM district_turnout dt
	WHERE dt.state NOT IN (SELECT state FROM irregular_state WHERE congress_nbr = $1)
	AND dt.congress_nbr = $1
	`
	q = fmt.Sprintf(q, stat)
	res, err := self.db.QueryContext(ctx, q, congress)
	if err != nil {
		panic(err)
	}
	defer res.Close()
	if !res.Next() {
		return 0, false
	}
	var avg *float64
	if err := res.Scan(&avg); err != nil {
		panic(err)
	}
	if avg == nil {
		return 0, false
	}
	return *avg, true
}

func (self *Db) MinVotersPerRegDistrict(ctx context.Context, congress int) (float64, bool) {
	return self._STATVotersPerRegDistrict(ctx, congress, "MIN")
}

func (self *Db) MaxVotersPerRegDistrict(ctx context.Context, congress int) (float64, bool) {
	return self._STATVotersPerRegDistrict(ctx, congress, "MAX")
}

func (self *Db) MeanVotersPerRegDistrict(ctx context.Context, congress int) (float64, bool) {
	return self._STATVotersPerRegDistrict(ctx, congress, "AVG")
}

type irregType struct {
	name  string
	table string
}

var gAtLargeAndNonAtLarge = irregType{
	name:  "Both at-large and non-at-large districts",
	table: "state_with_atlarge_and_nonatlarge_districts",
}

var gOverlappingTerms = irregType{
	name:  "Districts have multiple reps",
	table: "state_with_overlapping_terms",
}

var gIrregTypes = []irregType{gAtLargeAndNonAtLarge, gOverlappingTerms}

func stateIsIrregular(db *sql.DB, state string, congressNbr int, it irregType) (bool, error) {
	/*
		Uniform Congressional District Act was passed in 1967 (90th Congress).
		It exempted New Mexico and Hawaii. (Source: http://centerforpolitics.org/crystalball/articles/multi-member-legislative-districts-just-a-thing-of-the-past/)

		90th Congress: New Mexico and Hawaii had two at-large reps
		91th Congress: Hawaii had two at-large reps

		After that, all districts were regular. (Source: Wikipedia)
	*/

	if congressNbr == 90 {
		if it == gOverlappingTerms && (state == "NM" || state == "HI") {
			return true, nil
		}
		return false, nil
	}
	if congressNbr == 91 {
		if it == gOverlappingTerms && state == "HI" {
			return true, nil
		}
		return false, nil
	}
	if congressNbr >= 92 {
		return false, nil
	}

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
