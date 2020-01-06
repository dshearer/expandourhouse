package featureProc

import (
	"bufio"
	"encoding/json"
	"io"

	"github.com/paulmach/orb/geojson"
)

type ChoppedFeatureReader struct {
	r               *bufio.Reader
	bufferedFeature *geojson.Feature
}

func NewChoppedFeatureReader(r io.Reader) *ChoppedFeatureReader {
	return &ChoppedFeatureReader{r: bufio.NewReader(r), bufferedFeature: nil}
}

func (r *ChoppedFeatureReader) readInternal() (*geojson.Feature, error) {
	var line string
	var err error
	var f geojson.Feature

do:
	// read line
	line, err = r.r.ReadString('\n')
	// os.Stderr.WriteString(fmt.Sprintf("Read %v chars\n", len(line)))
	if len(line) == 0 && err == nil {
		goto do
	} else if len(line) == 0 {
		return nil, err
	}

	// parse it
	if err = json.Unmarshal([]byte(line), &f); err != nil {
		return nil, err
	}

	// return geojson.Feature
	return &f, nil
}

func (r *ChoppedFeatureReader) Read() (*geojson.Feature, bool, error) {
	var nextFeature *geojson.Feature
	var err error

	if r.bufferedFeature == nil {
		nextFeature, err = r.readInternal()
		if err != nil {
			return nil, false, err
		}
	} else {
		nextFeature = r.bufferedFeature
		r.bufferedFeature = nil
	}

	r.bufferedFeature, err = r.readInternal()
	if err == io.EOF {
		return nextFeature, true, nil
	}
	return nextFeature, false, err
}
