package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"expandourhouse.com/mapdata/housedb"
	"expandourhouse.com/mapdata/utils"
	"github.com/paulmach/orb/geojson"
	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

const gUsage = "usage: mark-irregular-states CONGRESS_NBR\n"

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
	var isIrregular bool
	var err error
	var rows *sql.Rows
	var nbr int

	sql := fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE state = $1 AND congress_nbr = $2", it.table)
	rows, err = db.Query(sql, state, congressNbr)
	if err != nil {
		goto done
	}
	rows.Next()
	if err = rows.Scan(&nbr); err != nil {
		goto done
	}
	isIrregular = nbr > 0

done:
	if rows != nil {
		rows.Close()
	}
	return isIrregular, err
}

func main() {
	log.SetOutput(os.Stderr)

	// parse args
	flag.Parse()
	if flag.NArg() != 1 {
		os.Stderr.WriteString(gUsage)
		os.Exit(1)
	}
	congressNbr, err := strconv.Atoi(flag.Arg(0))
	if err != nil || congressNbr <= 0 {
		os.Stderr.WriteString("Invalid Congress number\n")
		os.Exit(1)
	}

	// connect to DB
	var db *sql.DB
	for {
		db, err = GetData(context.Background())
		if err == nil {
			break
		}
		if !housedb.ErrIsDbLocked(err) {
			log.Fatal(err)
		}
		time.Sleep(3 * time.Second)
	}

	// find irregular states
	strm := stream.New(utils.NewFeatureReader(os.Stdin))
	strm.
		Map(func(f *geojson.Feature) *geojson.Feature {
			stateAbbr := f.Properties["state"].(string)

			/*
				All Congresses since the 90th are regular.
			*/
			if congressNbr > 90 {
				f.Properties["irregular"] = false
				return f
			}

			var irregHow []string
			for _, ir := range gIrregTypes {
				irreg, err := stateIsIrregular(
					db,
					stateAbbr,
					congressNbr,
					ir)
				if err != nil {
					log.Fatal(err)
				}
				if irreg {
					irregHow = append(irregHow, ir.name)
				}
			}

			f.Properties["irregular"] = len(irregHow) > 0
			if len(irregHow) > 0 {
				f.Properties["irregularHow"] = irregHow
			}
			return f
		}).
		// write to stdout
		Into(collectors.Func(func(data interface{}) error {
			f := data.(*geojson.Feature)
			encoder := json.NewEncoder(os.Stdout)
			return encoder.Encode(f)
		}))

	if err := <-strm.Open(); err != nil {
		log.Panic(err)
	}
}
