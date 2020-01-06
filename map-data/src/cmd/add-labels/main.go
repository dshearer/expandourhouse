package main

import (
	"flag"
	"log"
	"os"

	"expandourhouse.com/mapdata/featureProc"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
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

	proc := featureProc.NewWithReader(os.Stdin, os.Stdout).
		FlatMap(func(f *geojson.Feature) ([]*geojson.Feature, error) {
			if f.Geometry == nil || f.Properties.MustString("group") != "boundary" {
				return []*geojson.Feature{f}, nil
			}

			// make label
			labelPoint, _ := planar.CentroidArea(f.Geometry)
			labelFeature := geojson.NewFeature(labelPoint)
			labelFeature.Properties = f.Properties.Clone()
			labelFeature.Properties["group"] = "label"
			return []*geojson.Feature{f, labelFeature}, nil
		})
	if err := proc.Run(); err != nil {
		log.Panic(err)
	}
}
