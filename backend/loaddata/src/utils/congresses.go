package utils

import (
	"context"
	"database/sql"
	"fmt"
)

// GetCongressNbr returns the nbr for the congress that starts in the given
// year.
func GetCongressNbr(ctx context.Context, db *sql.DB, startYear int) (int, error) {
	var nbr int
	var err error
	var rows *sql.Rows

	sql := "SELECT nbr FROM congress WHERE start_year = $1"
	rows, err = db.QueryContext(ctx, sql, startYear)
	if err != nil {
		goto done
	}
	if !rows.Next() {
		err = fmt.Errorf("Couldn't find Congress for year %v", startYear)
		goto done
	}
	if err = rows.Scan(&nbr); err != nil {
		goto done
	}

done:
	if rows != nil {
		rows.Close()
	}
	return nbr, err
}
