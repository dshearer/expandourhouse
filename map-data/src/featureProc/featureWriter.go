package featureProc

import (
	"encoding/json"
	"io"
)

type ChoppedFeatureWriter = json.Encoder

func NewChoppedFeatureWriter(w io.Writer) *ChoppedFeatureWriter {
	return json.NewEncoder(w)
}
