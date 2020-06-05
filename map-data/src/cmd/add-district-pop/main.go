//go:generate go run scripts/includeData.go

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"expandourhouse.com/mapdata/housedb"
	"expandourhouse.com/mapdata/utils"
	"github.com/paulmach/orb/geojson"
	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

func getTurnout(db *sql.DB, congressNbr int, state string, districtNbr int) (*int, error) {
	var err error
	var rows *sql.Rows
	var turnout int
	var turnoutP = &turnout

	sql := "SELECT turnout FROM district_turnout WHERE state = $1 AND congress_nbr = $2 AND district_nbr = $3"
	rows, err = db.Query(sql, state, congressNbr, districtNbr)
	if err != nil {
		goto done
	}
	if !rows.Next() {
		turnoutP = nil
		goto done
	}
	if err = rows.Scan(&turnout); err != nil {
		turnoutP = nil
		goto done
	}

done:
	if rows != nil {
		rows.Close()
	}
	return turnoutP, err
}

func main() {
	log.SetOutput(os.Stderr)

	// parse args
	flag.Parse()
	if flag.NArg() != 0 {
		os.Stderr.WriteString("add-district-pop\n")
		os.Exit(1)
	}

	// connect to DB
	var db *sql.DB
	var err error
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

	strm := stream.New(utils.NewFeatureReader(os.Stdin))
	strm.
		Map(func(f *geojson.Feature) *geojson.Feature {
			if f.Properties["type"].(string) != "district" {
				return f
			}

			congressNbr := int(f.Properties["congress"].(float64))
			state := f.Properties["state"].(string)
			district := int(f.Properties["district"].(float64))

			// get turnout
			turnout, err := getTurnout(db, congressNbr, state, district)
			if err != nil {
				log.Fatal(err)
			}
			if turnout == nil {
				return f
			}
			os.Stderr.WriteString(fmt.Sprintf("Got turnout for %v-%v (%v)\n", state, district, congressNbr))
			f.Properties["turnout"] = *turnout
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
