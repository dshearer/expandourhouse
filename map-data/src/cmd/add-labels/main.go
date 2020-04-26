package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"expandourhouse.com/mapdata/utils"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

const gBoundaryGroup = "boundary"
const gLabelGroup = "label"

const gUsage = "usage: add-labels\n"

func main() {
	log.SetOutput(os.Stderr)

	// get args
	flag.Parse()
	if flag.NArg() != 0 {
		os.Stderr.WriteString(gUsage)
		os.Exit(1)
	}

	strm := stream.New(utils.NewFeatureReader(os.Stdin))

	strm.
		FlatMap(func(f *geojson.Feature) []*geojson.Feature {
			if f.Geometry == nil || f.Properties.MustString("group") != "boundary" {
				return []*geojson.Feature{f}
			}

			// make label
			labelPoint, _ := planar.CentroidArea(f.Geometry)
			labelFeature := geojson.NewFeature(labelPoint)
			labelFeature.Properties = f.Properties.Clone()
			labelFeature.Properties["group"] = "label"
			return []*geojson.Feature{f, labelFeature}
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
