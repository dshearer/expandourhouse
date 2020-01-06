package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"expandourhouse.com/mapdata/featureProc"
	"expandourhouse.com/mapdata/states"
	"github.com/paulmach/orb/geojson"
)

func intToOrdinal(i int) string {
	if i < 0 {
		panic("Can't handle negative int")
	}

	suffixes := []string{"th", "st", "nd", "rd", "th", "th", "th", "th", "th", "th"}
	var suffix string
	if i%100 == 11 || i%100 == 12 || i%100 == 13 {
		suffix = suffixes[0]
	} else {
		suffix = suffixes[i%10]
	}
	return fmt.Sprintf("%v%v", i, suffix)
}

func cleanUpProps(f *geojson.Feature) (*geojson.Feature, error) {
	// parse ID
	id, ok := f.Properties["ID"].(string)
	if !ok {
		return nil, errors.New("Feature doesn't have ID")
	}
	stateFips, err := strconv.Atoi(id[0:3])
	if err != nil {
		return nil, err
	}
	congress, err := strconv.Atoi(id[3:6])
	if err != nil {
		return nil, err
	}
	district, err := strconv.Atoi(id[10:12])
	if err != nil {
		return nil, err
	}

	// clean up feature's properties
	f.Properties = map[string]interface{}{
		"id":        id,
		"district":  district,
		"congress":  congress,
		"stateFips": stateFips,
		"state":     states.ByFips[stateFips].Usps,
		"group":     "boundary",
	}

	return f, nil
}

func addTitles(f *geojson.Feature) (*geojson.Feature, error) {
	district := f.Properties["district"].(int)
	var titleShortNbr string
	if district == 0 {
		titleShortNbr = "At Large"
	} else {
		titleShortNbr = fmt.Sprintf("%v", district)
	}
	var titleLongNbr string
	if district == 0 {
		titleLongNbr = "At Large"
	} else {
		titleLongNbr = intToOrdinal(district)
	}
	state := states.ByFips[f.Properties["stateFips"].(int)]

	f.Properties["titleShort"] = fmt.Sprintf("%v %v", state.Usps, titleShortNbr)
	f.Properties["titleLong"] = fmt.Sprintf("%v's %v Congressional District",
		state.Name, titleLongNbr)
	return f, nil
}

func main() {
	log.SetOutput(os.Stderr)

	flag.Parse()
	if flag.NArg() != 0 {
		os.Stderr.WriteString("usage: process-districts\n")
		os.Exit(1)
	}

	proc := featureProc.NewWithReader(os.Stdin, os.Stdout).
		Map(cleanUpProps).
		Filter(func(f *geojson.Feature) (bool, error) {
			// filter out invalid districts (e.g., -1 for Indian lands)
			return f.Properties["district"].(int) >= 0, nil
		}).
		Map(addTitles)

	if err := proc.Run(); err != nil {
		log.Panic(err)
	}
}
