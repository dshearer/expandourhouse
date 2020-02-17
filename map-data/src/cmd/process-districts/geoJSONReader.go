package main

import (
	"context"
	"encoding/json"
	"io"
	"log"

	"github.com/paulmach/orb/geojson"
)

type geoJSONReader struct {
	decoder *json.Decoder
	ctx     context.Context
	outChan chan interface{}
}

func newGeoJSONReader(r io.Reader) *geoJSONReader {
	var reader geoJSONReader
	reader.decoder = json.NewDecoder(r)
	reader.outChan = make(chan interface{})
	return &reader
}

func (r *geoJSONReader) read() {
	defer close(r.outChan)

	// parse json
	var collect geojson.FeatureCollection
	if err := r.decoder.Decode(&collect); err != nil {
		log.Panic(err)
	}

	features := collect.Features
	for len(features) > 0 {
		f := features[0]
		select {
		case r.outChan <- f:
			features = features[1:]
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *geoJSONReader) Open(ctx context.Context) error {
	r.ctx = ctx
	go r.read()
	return nil
}

func (r *geoJSONReader) GetOutput() <-chan interface{} {
	return r.outChan
}
