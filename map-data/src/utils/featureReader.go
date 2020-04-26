package utils

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"

	"github.com/paulmach/orb/geojson"
)

type FeatureReader struct {
	r       *bufio.Reader
	ctx     context.Context
	outChan chan interface{}
}

func NewFeatureReader(r io.Reader) *FeatureReader {
	var reader FeatureReader
	reader.r = bufio.NewReader(r)
	reader.outChan = make(chan interface{})
	return &reader
}

func (r *FeatureReader) read() {
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

func (r *FeatureReader) Open(ctx context.Context) error {
	r.ctx = ctx
	go r.read()
	return nil
}

func (r *FeatureReader) GetOutput() <-chan interface{} {
	return r.outChan
}
