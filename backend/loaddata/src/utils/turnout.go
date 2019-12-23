package utils

import (
	"context"
	"database/sql"
	"fmt"
)

var gDistrictHasTurnoutCache = make(map[string]bool)

func districtHasTurnoutCacheKey(districtId, sourceId int) string {
	return fmt.Sprintf("%5d%5d", districtId, sourceId)
}

func DistrictHasTurnout(ctx context.Context, db *sql.DB, districtId, sourceId int) (bool, error) {
	var hasTurnout bool
	var rows *sql.Rows
	var err error
	var sql string

	// check cache
	cacheKey := districtHasTurnoutCacheKey(districtId, sourceId)
	hasTurnout, gotFromCache := gDistrictHasTurnoutCache[cacheKey]
	if gotFromCache {
		goto done
	}

	// check DB
	sql = "SELECT COUNT(*) > 0 FROM house_district_turnout " +
		"WHERE house_district_id = $1 AND source_id = $2"
	rows, err = db.QueryContext(ctx, sql, districtId, sourceId)
	if err != nil {
		goto done
	}
	rows.Next()
	err = rows.Scan(&hasTurnout)

done:
	if rows != nil {
		rows.Close()
	}
	if !gotFromCache && err == nil {
		gDistrictHasTurnoutCache[cacheKey] = hasTurnout
	}
	return hasTurnout, err
}

func AddDistrictTurnout(ctx context.Context, db *sql.DB, districtId, sourceId,
	turnout int) error {

	hasTurnout, err := DistrictHasTurnout(ctx, db, districtId, sourceId)
	if err != nil {
		return err
	}

	if hasTurnout {
		sql := `UPDATE house_district_turnout SET num_votes = $1
		WHERE house_district_id = $2 AND source_id = $3`
		_, err = db.ExecContext(ctx, sql, turnout, districtId, sourceId)
	} else {
		sql := `INSERT INTO house_district_turnout(num_votes, house_district_id, source_id)
		VALUES($1, $2, $3)`
		_, err = db.ExecContext(ctx, sql, turnout, districtId, sourceId)
	}
	if err != nil {
		return err
	}

	// update cache
	cacheKey := districtHasTurnoutCacheKey(districtId, sourceId)
	gDistrictHasTurnoutCache[cacheKey] = true
	return nil
}
