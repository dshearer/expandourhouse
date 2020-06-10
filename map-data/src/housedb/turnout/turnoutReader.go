package turnout

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"expandourhouse.com/mapdata/housedb/sourceinst"
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
	cols      []string
}

func newTurnoutReader(source *sourceinst.SourceInst, comma rune) *turnoutReader {
	reader := csv.NewReader(source.Data)
	reader.Comma = comma
	reader.ReuseRecord = true
	return &turnoutReader{
		csvReader: reader,
		colToIdx:  nil,
	}
}

func (self *turnoutReader) IndexOfCol(col string) int {
	self.readCols()
	idx, ok := self.colToIdx[col]
	if !ok {
		panic(fmt.Sprintf("Unknown column: %v", col))
	}
	return idx
}

func (self *turnoutReader) readCols() {
	if self.colToIdx != nil {
		return
	}

	// get next record
	rec, err := self.csvReader.Read()
	if err != nil {
		panic(err)
	}

	// read cols
	self.colToIdx = make(map[string]int)
	for idx, col := range rec {
		col = strings.TrimSpace(col)
		self.colToIdx[col] = idx
		self.cols = append(self.cols, col)
	}
}

func (self *turnoutReader) Cols() []string {
	self.readCols()
	return self.cols
}

func (self *turnoutReader) Read() *turnoutRec {
	self.readCols()

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
