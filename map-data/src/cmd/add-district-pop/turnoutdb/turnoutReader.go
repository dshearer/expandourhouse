package turnoutdb

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"expandourhouse.com/mapdata/housedb"
)

type turnoutRec struct {
	Data     []string
	colToIdx map[string]int
}

func (self *turnoutRec) Get(col string) string {
	idx, ok := self.colToIdx[col]
	if !ok {
		panic(fmt.Sprintf("Unknown column: %v", col))
	}
	return self.Data[idx]
}

func (self *turnoutRec) GetInt(col string) int {
	s := self.Get(col)
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

type turnoutReader struct {
	csvReader *csv.Reader
	colToIdx  map[string]int
}

func newTurnoutReader(source *housedb.SourceInst, comma rune) *turnoutReader {
	reader := csv.NewReader(source.Data)
	reader.Comma = comma
	reader.ReuseRecord = true
	return &turnoutReader{
		csvReader: reader,
		colToIdx:  nil,
	}
}

func (self *turnoutReader) IndexOfCol(col string) int {
	idx, ok := self.colToIdx[col]
	if !ok {
		panic(fmt.Sprintf("Unknown column: %v", col))
	}
	return idx
}

func (self *turnoutReader) readCols() {
	// get next record
	rec, err := self.csvReader.Read()
	if err != nil {
		panic(err)
	}

	// read cols
	self.colToIdx = make(map[string]int)
	for idx, col := range rec {
		self.colToIdx[strings.TrimSpace(col)] = idx
	}
}

func (self *turnoutReader) Cols() []string {
	if self.colToIdx == nil {
		self.readCols()
	}
	cols := make([]string, len(self.colToIdx))
	for col := range self.colToIdx {
		cols = append(cols, col)
	}
	return cols
}

func (self *turnoutReader) Read() *turnoutRec {
	if self.colToIdx == nil {
		self.readCols()
	}

	// get next record
	rec, err := self.csvReader.Read()
	if err == io.EOF {
		return nil
	} else if err != nil {
		panic(err)
	}

	var newRec turnoutRec
	newRec.Data = make([]string, len(rec))
	newRec.colToIdx = self.colToIdx
	for idx, val := range rec {
		newRec.Data[idx] = strings.TrimSpace(val)
	}
	return &newRec
}
