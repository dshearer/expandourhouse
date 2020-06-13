package turnout

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"expandourhouse.com/mapdata/bulkInserter"
	"expandourhouse.com/mapdata/congresses"
	"expandourhouse.com/mapdata/housedb/sourceinst"
	"expandourhouse.com/mapdata/states"
	_ "github.com/mattn/go-sqlite3"
)

var gWordsToNumbers = map[string]int{
	"one":          1,
	"two":          2,
	"three":        3,
	"four":         4,
	"five":         5,
	"six":          6,
	"seven":        7,
	"eight":        8,
	"nine":         9,
	"ten":          10,
	"eleven":       11,
	"twelve":       12,
	"thirteen":     13,
	"fourteen":     14,
	"fifteen":      15,
	"sixteen":      16,
	"seventeen":    17,
	"eighteen":     18,
	"nineteen":     19,
	"twenty":       20,
	"twenty one":   21,
	"twenty-one":   21,
	"twenty two":   22,
	"twenty three": 23,
	"twenty four":  24,
	"twenty five":  25,
	"twenty six":   26,
	"twenty seven": 27,
	"twenty eight": 28,
	"twenty nine":  29,
	"thirty":       30,
}

var gLocalityCols = []string{"Territory", "City", "County", "Town", "Township",
	"Ward", "Parish", "Populated Place", "Hundred", "Borough"}

func isEmptyValue(v *string) bool {
	return v == nil || *v == "null"
}

/*
	Each record contains the number of votes for a specific candidate
	in a specific state and district and, maybe, in a specific locality
	within the district. The records without a locality contain the
	total votes for a candidate for the district.
*/

/*
	The records for this ID *should* have at most one unique value
	for "district". "0" means that it is an at-large
	district.

	If there is one unique non-"0" value along with "0",
	then the records with "0" are redundant.

	If there are multiple unique non-"0" values, then we just ignore
	all these records.
*/

func parseYear(s string) (int, bool) {
	dashIdx := strings.IndexRune(s, '-')
	if dashIdx != -1 {
		s = s[:dashIdx]
	}
	year, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return year, true
}

func parseDistrict(s *string) int {
	if s == nil {
		return 0
	}
	district, ok := gWordsToNumbers[strings.ToLower(*s)]
	if !ok {
		return -1
	}
	return district
}

type rawDb struct {
	db  *sql.DB
	dir string
}

func (self *rawDb) Close() {
	if self.db == nil {
		return
	}
	self.db.Close()
	os.RemoveAll(self.dir)
	self.db = nil
}

func makeTuftsRawDb(ctx context.Context, source *sourceinst.SourceInst) (*rawDb, error) {
	// make temp DB
	tmpDbDir, err := ioutil.TempDir("", "tuftsDb*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDbDir)
	tmpDb, err := sql.Open("sqlite3", path.Join(tmpDbDir, "db"))
	if err != nil {
		panic(err)
	}

	// make reader
	reader := newTurnoutReader(source, '\t')

	// read columns
	var tableCols []string
	var tableColsWithType []string
	for _, col := range reader.Cols() {
		tableCol := strings.ReplaceAll(col, " ", "_")
		typ := "TEXT"
		if tableCol == "Vote" || tableCol == "District" || tableCol == "Date" {
			typ = "INT"
		}
		tableCols = append(tableCols, tableCol)
		tableColsWithType = append(tableColsWithType, fmt.Sprintf("%s %v", tableCol, typ))
	}

	// make temp table
	makeTableQuery := fmt.Sprintf(
		"CREATE TABLE raw_tufts(%v)",
		strings.Join(tableColsWithType, ","),
	)
	_, err = tmpDb.ExecContext(ctx, makeTableQuery)
	if err != nil {
		panic(err)
	}

	// insert TSV data
	tx, err := tmpDb.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}
	inserter := bulkInserter.Make(ctx, tx, "raw_tufts", tableCols)
	inserter.FlushPeriod = 25
	skippedBcOfType := 0
	skippedBcOfLocality := 0
	inserted := 0
	for {
		// get record
		rec := reader.Read()
		if rec == nil {
			break
		}

		// make sure it's for a general election
		if *rec.Get("Type") != "General" {
			skippedBcOfType++
			continue
		}

		// make sure it's for the House
		if *rec.Get("Role") != "U.S. Congressman" {
			continue
		}

		idLower := strings.ToLower(*rec.Get("id"))
		if strings.Contains(idLower, "district") {
			/*
				This is evidence that this district has a
				weird name like "Eastern".
			*/
			continue
		}
		if strings.Contains(idLower, "congressman") ||
			strings.Contains(idLower, "congressmen") ||
			strings.Contains(idLower, "1strep") ||
			strings.Contains(idLower, "2ndrep") {
			/* This is evidence of multi-representative district */
			continue
		}
		if strings.Contains(idLower, "special") {
			/* This is evidence of some stupid special election */
			continue
		}

		if *rec.Get("State") == "South Carolina" {
			/* This state's data is a mess */
			continue
		}

		// make sure it's for an entire district
		forEntireDistrict := true
		for _, col := range gLocalityCols {
			if !isEmptyValue(rec.Get(col)) {
				forEntireDistrict = false
				break
			}
		}
		if !forEntireDistrict {
			skippedBcOfLocality++
			continue
		}

		// parse district
		rec.Set("District", fmt.Sprintf("%v", parseDistrict(rec.Get("District"))))

		// parse state
		state, ok := states.ByName[*rec.Get("State")]
		if !ok {
			continue
		}
		rec.Set("State", state.Usps)

		// parse year
		year, ok := parseYear(*rec.Get("Date"))
		if !ok {
			continue
		}
		rec.Set("Date", fmt.Sprintf("%v", year))

		// parse vote
		if rec.Get("Vote") != nil {
			vote, err := strconv.Atoi(*rec.Get("Vote"))
			if err != nil {
				continue
			}
			rec.Set("Vote", fmt.Sprintf("%v", vote))
		}

		// insert into DB
		var tmp []interface{}
		for _, v := range rec.Data {
			tmp = append(tmp, v)
		}
		if err := inserter.Insert(tmp); err != nil {
			panic(err)
		}
		inserted++
	}

	log.Printf("skippedBcOfType: %v\n", skippedBcOfType)
	log.Printf("skippedBcOfLocality: %v\n", skippedBcOfLocality)
	log.Printf("inserted: %v\n", inserted)

	// flush to DB
	if err := inserter.Flush(); err != nil {
		panic(err)
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}

	return &rawDb{db: tmpDb, dir: tmpDbDir}, nil
}

/*
	ID formats:

	al.congress.1819
		- No districts

	al.uscongress.1.1823
		- For 1st district

	ct.special.congress.1790
		- Special election

	ct.special2.congress.1793
		- Special election

	ct.uscongress.special.1.1801
		- Special election

	ct.congress.special.1805
		- Special election

	ga.uscongress.NorthernDistrict.1791
		- For Northern district

	ga.specialuscongress1.1806

	ky.uscongress3.1812
		- For third district

	ky.uscongress5.special.1810
	ky.specialuscongress1.1816
	me.uscongress4.secondrunoff.1821

	me.uscongress.3.2.1821
		- 2nd ballot
		- 3rd district

	me.uscongress3.third.1823
		- 3rd ballot
		- 3rd district

	md.uscongress5.special.18thcongress.1823

	ma.uscongress.2.hampshire.ballot2.1793
		- 2nd ballot
		- 2nd district

	ma.uscongress.eastern.2.1798
		- 1st ballot
		- Eastern 2 district

	ny.uscongressspecial.1791
*/

func getDistrictFromId(id string) int {
	congress := "congress"
	uscongress := "uscongress"
	parts := strings.Split(id, ".")
	part1 := parts[1]
	if !strings.HasPrefix(part1, congress) &&
		!strings.HasPrefix(part1, uscongress) {
		return -1
	}

	if strings.HasPrefix(part1, congress) {
		part1 = part1[len(congress):]
	} else if strings.HasPrefix(part1, uscongress) {
		part1 = part1[len(uscongress):]
	}
	var d int
	if len(part1) == 0 {
		var err error
		d, err = strconv.Atoi(parts[2])
		if err != nil {
			return -1
		}
	} else {
		var err error
		d, err = strconv.Atoi(part1)
		if err != nil {
			return -1
		}
	}
	if d > 100 {
		return -1
	}
	return d
}

func addTuftsData(ctx context.Context, tx *sql.Tx, source *sourceinst.SourceInst) error {
	/*
		1. Read records from TSV file into a temporary Sqlite DB
		2. Use the temp DB to compute turnouts
		3. Insert turnouts into DB
	*/

	// read records into temp DB
	rawDb, err := makeTuftsRawDb(ctx, source)
	if err != nil {
		panic(err)
	}
	defer rawDb.Close()

	/*
		Each ID corresponds to a set of records that belong to one district
		and one election
	*/
	// get unique rec IDs
	res, err := rawDb.db.QueryContext(ctx, "SELECT DISTINCT id FROM raw_tufts")
	if err != nil {
		return err
	}
	var recIds []string
	for res.Next() {
		var id string
		if err := res.Scan(&id); err != nil {
			res.Close()
			return err
		}
		recIds = append(recIds, id)
	}
	res.Close()

	// make bulk-inserter
	inserter := bulkInserter.Make(ctx, tx, gTableName, gTableCols[:])

	for _, id := range recIds {
		// get unique districts
		res, err = rawDb.db.QueryContext(ctx,
			"SELECT DISTINCT district FROM raw_tufts WHERE id = ?", id)
		if err != nil {
			return err
		}
		districts := make(map[int]bool)
		for res.Next() {
			var d int
			if err := res.Scan(&d); err != nil {
				res.Close()
				return err
			}
			districts[d] = true
		}
		res.Close()

		// check for invalid district
		if _, ok := districts[-1]; ok {
			continue
		}

		/*
			Sometimes even when the id is for a specific district,
			the record set includes (duplicate) records with district == 0.
		*/
		if len(districts) > 1 {
			/* What do we do with this? */
			continue
		}

		// decide the actual district in question
		actualDistrict := 0
		for d := range districts {
			if d != 0 {
				actualDistrict = d
				break
			}
		}

		// see if this agrees with the district in the ID
		d := getDistrictFromId(id)
		if d != -1 && d != actualDistrict {
			continue
		}

		/*
			Sometimes record sets are just bad.
		*/

		// check for multiple records for same candidate
		q := `
		WITH t AS (
			SELECT COUNT(*) cnt FROM raw_tufts
			WHERE id = ? AND District = ?
			GROUP BY Name_ID)
		SELECT * FROM t WHERE cnt > 1`
		res, err = rawDb.db.QueryContext(ctx, q, id, actualDistrict)
		if err != nil {
			return err
		}
		if res.Next() {
			res.Close()
			continue
		}
		res.Close()

		// check for missing vote counts
		q = `SELECT * FROM raw_tufts WHERE id = ? AND vote IS NULL`
		res, err = rawDb.db.QueryContext(ctx, q, id)
		if err != nil {
			return err
		}
		if res.Next() {
			res.Close()
			continue
		}
		res.Close()

		// get year and state
		res, err = rawDb.db.QueryContext(ctx,
			"SELECT Date, State FROM raw_tufts WHERE id = ?", id)
		if err != nil {
			return err
		}
		if !res.Next() {
			continue
		}
		var year int
		var state string
		if err := res.Scan(&year, &state); err != nil {
			res.Close()
			return err
		}
		res.Close()

		if year%2 == 1 {
			/* This avoids bogus dup entries for the same congress */
			continue
		}

		// get congress
		/*
			The election was either the year before the session of Congress,
			or in the same year.
		*/
		if year%2 == 0 {
			year++
		}
		congress := congresses.GetForYear(year)
		if congress == nil {
			log.Printf("Can't find congress for %v (year %v)\n", id, year)
			continue
		}

		// get total votes
		countSql := "SELECT SUM(Vote) FROM raw_tufts WHERE id = ? AND District = ?"
		res, err = rawDb.db.QueryContext(ctx, countSql, id, actualDistrict)
		if err != nil {
			return err
		}
		if !res.Next() {
			log.Printf("countSql returned zero records for %v\n", id)
			res.Close()
			continue
		}
		var totalVotes int
		if err := res.Scan(&totalVotes); err != nil {
			res.Close()
			return err
		}
		res.Close()

		// insert total votes into DB
		values := []interface{}{
			actualDistrict,
			state,
			congress.Number,
			totalVotes,
		}
		if err := inserter.Insert(values); err != nil {
			return err
		}
	} // for

	// finish up
	if err = inserter.Flush(); err != nil {
		return err
	}

	return nil
}
