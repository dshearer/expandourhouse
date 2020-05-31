package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"expandourhouse.com/mapdata/states"
	"expandourhouse.com/mapdata/utils"
	"github.com/paulmach/orb/geojson"
	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

func cleanUpProps(f *geojson.Feature) *geojson.Feature {
	// parse ID
	id, ok := f.Properties["ID"].(string)
	if !ok {
		log.Panic("Feature doesn't have ID")
	}
	stateFips, err := strconv.Atoi(id[0:3])
	if err != nil {
		log.Panic(err)
	}
	congress, err := strconv.Atoi(id[3:6])
	if err != nil {
		log.Panic(err)
	}
	district, err := strconv.Atoi(id[10:12])
	if err != nil {
		log.Panic(err)
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

	return f
}

func addTitles(f *geojson.Feature) *geojson.Feature {
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
		titleLongNbr = utils.IntToOrdinal(district)
	}
	state := states.ByFips[f.Properties["stateFips"].(int)]

	f.Properties["titleShort"] = fmt.Sprintf("%v %v", state.Usps, titleShortNbr)
	f.Properties["titleLong"] = fmt.Sprintf("%v District", titleLongNbr)
	return f
}

func main() {
	log.SetOutput(os.Stderr)

	flag.Parse()
	if flag.NArg() != 0 {
		os.Stderr.WriteString("usage: process-districts\n")
		os.Exit(1)
	}

	strm := stream.New(newGeoJSONReader(os.Stdin))
	strm.
		Map(cleanUpProps).
		Filter(func(f *geojson.Feature) bool {
			// filter out invalid districts (e.g., -1 for Indian lands)
			return f.Properties["district"].(int) >= 0
		}).
		Map(addTitles).
		Into(collectors.Func(func(data interface{}) error {
			f := data.(*geojson.Feature)
			encoder := json.NewEncoder(os.Stdout)
			return encoder.Encode(f)
		}))

	if err := <-strm.Open(); err != nil {
		log.Panic(err)
	}
}
