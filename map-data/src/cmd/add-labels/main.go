package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"

	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

const gBoundaryGroup = "boundary"
const gLabelGroup = "label"

const gUsage = "usage: add-labels\n"

type featureReader struct {
	r       *bufio.Reader
	ctx     context.Context
	outChan chan interface{}
}

func newFeatureReader(r io.Reader) *featureReader {
	var reader featureReader
	reader.r = bufio.NewReader(r)
	reader.outChan = make(chan interface{})
	return &reader
}

func (r *featureReader) read() {
	defer close(r.outChan)

	for {
		var line string
		var err error
		var f geojson.Feature

		// read line
		line, err = r.r.ReadString('\n')
		if len(line) == 0 && err != nil {
			if err == io.EOF {
				return
			}
			log.Panic(err)
		}

		if len(line) == 0 {
			continue
		}

		// parse it
		if err = json.Unmarshal([]byte(line), &f); err != nil {
			log.Panic(err)
		}

		// send it
		select {
		case r.outChan <- &f:
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *featureReader) Open(ctx context.Context) error {
	r.ctx = ctx
	go r.read()
	return nil
}

func (r *featureReader) GetOutput() <-chan interface{} {
	return r.outChan
}

func main() {
	log.SetOutput(os.Stderr)

	// get args
	flag.Parse()
	if flag.NArg() != 0 {
		os.Stderr.WriteString(gUsage)
		os.Exit(1)
	}

	strm := stream.New(newFeatureReader(os.Stdin))

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
