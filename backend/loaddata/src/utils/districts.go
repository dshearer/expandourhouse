package utils

import (
	"context"
	"database/sql"
	"fmt"
)

var gDistrictCache = make(map[string]int)

func districtCacheKey(state string, district, congressNbr int) string {
	return fmt.Sprintf("%v%2d%3d", state, district, congressNbr)
}

// GetDistrict returns the ID of the table row corresponding to the specified
// district.  If this district isn't in the DB yet, it is added.
func GetDistrict(ctx context.Context, db *sql.DB,
	state string, district int, congressNbr int) (int, error) {

	var rowId int
	var rows *sql.Rows
	var err error

	// check cache
	cacheKey := districtCacheKey(state, district, congressNbr)
	rowId, gotFromCache := gDistrictCache[cacheKey]
	if gotFromCache {
		goto done
	}

	// check if there's already a row for this district
	rows, err = db.QueryContext(ctx, "SELECT id FROM house_district WHERE "+
		"state = $1 AND district = $2 AND congress_nbr = $3",
		state, district, congressNbr)
	if err != nil {
		goto done
	}
	if rows.Next() {
		/* Already exists */
		err = rows.Scan(&rowId)
		goto done
	}

	// make district row
	rows.Close()
	rows, err = db.QueryContext(ctx, "INSERT INTO house_district(state, district, congress_nbr)"+
		" VALUES ($1, $2, $3) RETURNING id", state, district, congressNbr)
	if err != nil {
		goto done
	}
	rows.Next()
	err = rows.Scan(&rowId)

done:
	if rows != nil {
		rows.Close()
	}
	if !gotFromCache && err == nil {
		gDistrictCache[cacheKey] = rowId
	}
	return rowId, err
}
