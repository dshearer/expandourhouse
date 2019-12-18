package main

import (
	"context"
	"database/sql"
)

func getDistrict(ctx context.Context, db *sql.DB,
	state string, district int, congressNbr int) (int, error) {

	var rowId int
	var rows *sql.Rows
	var err error

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
	return rowId, err
}
