package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"

	"expandourhouse.com/mapdata/utils"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

const gUsage = "usage: reduce-precision NBR_OF_DECIMALS\n"

func main() {
	log.SetOutput(os.Stderr)

	// get args
	flag.Parse()
	if flag.NArg() != 1 {
		os.Stderr.WriteString(gUsage)
		os.Exit(1)
	}
	precision, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	if precision < 0 {
		os.Stderr.WriteString("NBR_OF_DECIMALS must not be negative\n")
		os.Exit(1)
	}

	// do it
	strm := stream.New(utils.NewFeatureReader(os.Stdin))
	strm.
		Map(func(f *geojson.Feature) *geojson.Feature {
			f.Geometry = orb.Round(f.Geometry, precision)
			return f
		}).
		Into(collectors.Func(func(data interface{}) error {
			f := data.(*geojson.Feature)
			encoder := json.NewEncoder(os.Stdout)
			return encoder.Encode(f)
		}))

	if err := <-strm.Open(); err != nil {
		log.Panic(err)
	}
}
