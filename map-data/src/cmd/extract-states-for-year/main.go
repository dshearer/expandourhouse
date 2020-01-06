package main

import (
	"encoding/json"
	"expandourhouse.com/mapdata/featureProc"
	"flag"
	"github.com/paulmach/orb/geojson"
	"log"
	"os"
	"strconv"
	"time"
)

const gUsage = "usage: extract-states-for-year HISTORICAL_STATES_FILE YEAR\n"

type collectionSupplier struct {
	collection *geojson.FeatureCollection
	pos        int
}

func (s *collectionSupplier) Get() (*geojson.Feature, error) {
	if s.pos >= len(s.collection.Features) {
		return nil, nil
	}

	obj := s.collection.Features[s.pos]
	s.pos++
	return obj, nil
}

func main() {
	log.SetOutput(os.Stderr)

	// get args
	flag.Parse()
	if flag.NArg() != 2 {
		os.Stderr.WriteString(gUsage)
		os.Exit(1)
	}
	histStatesFile := flag.Arg(0)
	year, err := strconv.Atoi(flag.Arg(1))
	if err != nil || year <= 0 {
		os.Stderr.WriteString("Invalid year\n")
		os.Exit(1)
	}

	// cap year at 1980, since the states haven't changed since then
	if year > 1980 {
		year = 1980
	}

	// parse historical-states file
	f, err := os.Open(histStatesFile)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()
	var features geojson.FeatureCollection
	decoder := json.NewDecoder(f)
	if err = decoder.Decode(&features); err != nil {
		log.Panic(err)
	}

	// make processor
	supplier := collectionSupplier{&features, 0}
	proc := featureProc.NewWithSupplier(&supplier, os.Stdout).
		// filter states by date
		Filter(func(f *geojson.Feature) (bool, error) {
			// parse dates
			startDateStr := f.Properties.MustString("START_DATE", "")
			endDateStr := f.Properties.MustString("END_DATE", "")
			if len(startDateStr) == 0 || len(endDateStr) == 0 {
				return false, nil
			}
			startDate, err := time.Parse("2006/01/02", startDateStr)
			if err != nil {
				return false, err
			}
			endDate, err := time.Parse("2006/01/02", endDateStr)
			if err != nil {
				return false, err
			}

			return startDate.Year() <= year && year <= endDate.Year(), nil
		}).
		// add some properties
		Map(func(f *geojson.Feature) (*geojson.Feature, error) {
			f.Properties["group"] = "boundary"
			f.Properties["id"] = f.Properties["ID"]
			delete(f.Properties, "ID")
			f.Properties["titleLong"] = f.Properties["FULL_NAME"]
			f.Properties["titleShort"] = f.Properties["ABBR_NAME"]
			return f, nil
		})

	if err := proc.Run(); err != nil {
		log.Panic(err)
	}
}
