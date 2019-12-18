package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type stateDataEntry struct {
	Name string
	FIPS int
	USPS string
}

// Returns a map from state FIPS code to state abbrev
func LoadStateData(dataDirPath string) (map[int]string, error) {
	statesFilePath := filepath.Join(dataDirPath, "states.json")
	f, err := os.Open(statesFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	states := make(map[int]string)
	if !decoder.More() {
		return nil, errors.New("States file seems to be empty")
	}
	var entries []stateDataEntry
	err = decoder.Decode(&entries)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		states[entry.FIPS] = entry.USPS
	}

	return states, nil
}
