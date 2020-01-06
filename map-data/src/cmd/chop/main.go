package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"expandourhouse.com/mapdata/featureProc"
	"github.com/paulmach/orb/geojson"
)

func chop() {
	// parse json
	decoder := json.NewDecoder(os.Stdin)
	var collect geojson.FeatureCollection
	if err := decoder.Decode(&collect); err != nil {
		log.Panic(err)
	}
	os.Stderr.WriteString(fmt.Sprintf("Read %v Features\n", len(collect.Features)))

	// chop the input
	writer := featureProc.NewChoppedFeatureWriter(os.Stdout)
	for _, f := range collect.Features {
		if err := writer.Encode(f); err != nil {
			log.Panic(err)
		}
	}
}

func unchop() {
	// write beginning
	os.Stdout.WriteString("{\n\"Features\":[\n")

	// write Features
	featureReader := featureProc.NewChoppedFeatureReader(os.Stdin)
	for {
		// read Feature
		f, isLast, err := featureReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Panic(err)
		}

		// write Feature
		featureBytes, err := json.Marshal(f)
		if err != nil {
			log.Panic(err)
		}
		if _, err = os.Stdout.Write(featureBytes); err != nil {
			log.Panic(err)
		}
		if !isLast {
			os.Stdout.WriteString(",")
		}
		os.Stdout.WriteString("\n")
	}

	// write end
	os.Stdout.WriteString("],\n\"type\": \"FeatureCollection\"\n}\n")
}

func main() {
	log.SetOutput(os.Stderr)

	shouldUnchop := flag.Bool("u", false, "unchop the input")
	flag.Parse()

	if *shouldUnchop {
		unchop()
	} else {
		chop()
	}
}
