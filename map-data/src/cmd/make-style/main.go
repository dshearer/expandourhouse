//go:generate go run scripts/includeStyle.go

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"expandourhouse.com/mapdata/stylemetadata"
)

var gUSAGE = "usage: makeStyle CONGRESS_NBR MAPBOX_USER\n"

func main() {
	log.SetOutput(os.Stderr)

	// get args
	flag.Parse()
	if flag.NArg() != 2 {
		os.Stderr.WriteString(gUSAGE)
		os.Exit(1)
	}
	congress, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		os.Stderr.WriteString(gUSAGE)
		os.Exit(1)
	}
	mapboxUser := flag.Arg(1)

	// make tileset IDs and names
	/* ID: max 32 characters, only one period */
	statesTilesetID := fmt.Sprintf("%v.states-%v", mapboxUser, congress)
	districtsTilesetID := fmt.Sprintf("%v.districts-%v", mapboxUser, congress)
	/* Name: max 64 characters, no spaces */
	statesTilesetName := fmt.Sprintf("US_States_%v", congress)
	districtsTilesetName := fmt.Sprintf("US_Districts_%v", congress)

	// make style
	var style map[string]interface{}
	if err = json.Unmarshal([]byte(gStyleTemplate), &style); err != nil {
		log.Panic(err)
	}
	style["name"] = fmt.Sprintf("congress-%v", congress)
	style["glyphs"] = fmt.Sprintf("mapbox://fonts/%v/{fontstack}/{range}.pbf",
		mapboxUser)
	sources := style["sources"].(map[string]interface{})
	composite := sources["composite"].(map[string]interface{})
	newURL := fmt.Sprintf("%s,%s,%s",
		composite["url"],
		statesTilesetID,
		districtsTilesetID)
	composite["url"] = newURL
	metadata := stylemetadata.StyleMetadata{
		StatesTilesetID:      statesTilesetID,
		StatesTilesetName:    statesTilesetName,
		DistrictsTilesetID:   districtsTilesetID,
		DistrictsTilesetName: districtsTilesetName,
	}
	stylemetadata.Set(style, mapboxUser, &metadata)

	// serialize it
	encoder := json.NewEncoder(os.Stdout)
	encoder.Encode(style)
}
