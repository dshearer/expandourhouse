package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"expandourhouse.com/mapdata/housedb"
	"expandourhouse.com/mapdata/utils"
	"github.com/paulmach/orb/geojson"
	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
	"golang.org/x/text/message"
)

func main() {
	log.SetOutput(os.Stderr)
	ctx := context.Background()

	// parse args
	flag.Parse()
	if flag.NArg() != 0 {
		os.Stderr.WriteString("add-district-pop\n")
		os.Exit(1)
	}

	// connect to DB
	db := housedb.Connect(context.Background())
	defer db.Close()

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
			turnout := db.GetTurnout(ctx, congressNbr, state, district)
			if turnout == nil {
				return f
			}
			os.Stderr.WriteString(fmt.Sprintf("Got turnout for %v-%v (%v)\n", state, district, congressNbr))
			p := message.NewPrinter(message.MatchLanguage("en"))
			f.Properties["turnout"] = p.Sprint(*turnout)
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
