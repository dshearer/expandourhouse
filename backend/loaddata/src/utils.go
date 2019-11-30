package main

import (
	"context"
	"database/sql"
)

func getSource(ctx context.Context, db *sql.DB, text string) (int, error) {
	var id int
	var err error
	var rows *sql.Rows

	rows, err = db.QueryContext(ctx, "SELECT id FROM source WHERE name = $1", text)
	if err != nil {
		goto done
	}
	if rows.Next() {
		err = rows.Scan(&id)
		goto done
	}

	// make source row
	rows.Close()
	rows, err = db.QueryContext(ctx, "INSERT INTO source(name) VALUES ($1) RETURNING id",
		text)
	if err != nil {
		goto done
	}
	rows.Next()
	err = rows.Scan(&id)

done:
	if rows != nil {
		rows.Close()
	}
	return id, err
}
