package utils

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type stateDataEntry struct {
	Name string
	FIPS int
	USPS string
}

var gEntries []*stateDataEntry
var gFipsToIdx map[int]int
var gNameToIdx map[string]int

func LoadStateData(dataDirPath string) error {
	if gEntries != nil {
		return nil
	}

	// open data file
	statesFilePath := filepath.Join(dataDirPath, "states.json")
	f, err := os.Open(statesFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// decode as json
	decoder := json.NewDecoder(f)
	if !decoder.More() {
		return errors.New("States file seems to be empty")
	}
	err = decoder.Decode(&gEntries)
	if err != nil {
		return err
	}

	// process entries
	gFipsToIdx = make(map[int]int)
	gNameToIdx = make(map[string]int)
	for idx, entry := range gEntries {
		gFipsToIdx[entry.FIPS] = idx
		gNameToIdx[strings.ToLower(entry.Name)] = idx
	}

	return nil
}

func GetUspsStateForFips(fipsCode int) (string, error) {
	if gEntries == nil {
		return "", errors.New("State data not loaded")
	}

	idx, ok := gFipsToIdx[fipsCode]
	if !ok {
		return "", errors.New("No such state")
	}
	return gEntries[idx].USPS, nil
}

func GetUspsStateForName(name string) (string, error) {
	if gEntries == nil {
		return "", errors.New("State data not loaded")
	}

	idx, ok := gNameToIdx[strings.ToLower(name)]
	if !ok {
		return "", errors.New("No such state")
	}
	return gEntries[idx].USPS, nil
}
