package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"

	"expandourhouse.com/mapdata/housedb"
	"expandourhouse.com/mapdata/utils"
	"github.com/paulmach/orb/geojson"
	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

const gUsage = "usage: mark-irregular-states CONGRESS_NBR\n"

func main() {
	log.SetOutput(os.Stderr)
	ctx := context.Background()

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
	db := housedb.Connect(ctx)
	defer db.Close()

	// find irregular states
	strm := stream.New(utils.NewFeatureReader(os.Stdin))
	strm.
		Map(func(f *geojson.Feature) *geojson.Feature {
			stateAbbr := f.Properties["state"].(string)
			irregHow := db.StateIrregularities(ctx, stateAbbr, congressNbr)
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
