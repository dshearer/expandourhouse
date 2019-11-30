package main

import (
	"bufio"
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"strings"
)

func UpdateCongresses(ctx context.Context, db *sql.DB, dataDirPath string) error {
	// open data file
	f, err := os.Open(dataDirPath + "/congress-start-years.txt")
	if err != nil {
		return err
	}
	defer f.Close()

	nextCongress := 1
	scanner := bufio.NewScanner(f)
	nbrInserted := 0
	for scanner.Scan() {
		// parse line
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		year, _ := strconv.Atoi(line)

		// check if we already have this congress
		sql := "SELECT COUNT(*) FROM congress WHERE nbr = $1 AND start_year = $2"
		row, err := db.QueryContext(ctx, sql, nextCongress, year)
		if err != nil {
			return err
		}
		row.Next()
		var n int
		row.Scan(&n)
		haveCongress := n > 0
		row.Close()

		if !haveCongress {
			// insert congress
			sql = "INSERT INTO congress(nbr, start_year) VALUES ($1, $2)"
			if _, err = db.ExecContext(ctx, sql, nextCongress, year); err != nil {
				return err
			}
			nbrInserted++
		}
		nextCongress++
	}
	log.Printf("Inserted %v congress sessions", nbrInserted)

	return nil
}
