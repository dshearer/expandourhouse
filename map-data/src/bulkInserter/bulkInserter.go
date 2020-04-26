package bulkInserter

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

func makeValuePlaceholder(startNbr int, endNbr int) string {
	var parts []string
	for i := startNbr; i < endNbr; i++ {
		parts = append(parts, fmt.Sprintf("$%v", i))
	}
	return fmt.Sprintf("(%v)", strings.Join(parts, ", "))
}

type Inserter struct {
	ctx      context.Context
	db       *sql.DB
	table    string
	numCols  int
	sqlStart string
	buffer   []interface{}
}

func Make(ctx context.Context, db *sql.DB, table string,
	cols []string) Inserter {

	var inserter Inserter
	inserter.ctx = ctx
	inserter.db = db
	inserter.table = table
	inserter.numCols = len(cols)

	// make parts of SQL statement
	inserter.sqlStart = fmt.Sprintf("INSERT INTO %v(%v) VALUES ", table, strings.Join(cols, ", "))
	var qs []string
	for i := 0; i < len(cols); i++ {
		qs = append(qs, "?")
	}

	return inserter
}

func (self *Inserter) Flush() error {
	if len(self.buffer) == 0 {
		return nil
	}

	// make SQL
	var placeholders []string
	n := 1
	for i := 0; i < len(self.buffer)/self.numCols; i++ {
		n2 := n + self.numCols
		placeholders = append(placeholders, makeValuePlaceholder(n, n2))
		n = n2
	}
	sql := fmt.Sprintf("%v %v", self.sqlStart, strings.Join(placeholders, ", "))

	// execute SQL
	_, err := self.db.ExecContext(self.ctx, sql, self.buffer...)
	if err != nil {
		return errors.Wrapf(err, "'%v' with %v params", sql, len(self.buffer))
	}

	// empty buffer
	self.buffer = nil

	return nil
}

func (self *Inserter) Insert(values []interface{}) error {
	if len(values) != self.numCols {
		panic(fmt.Sprintf("Bad number of values: got %v instead of %v",
			len(values), self.numCols))
	}

	// add values to buffer
	for _, val := range values {
		self.buffer = append(self.buffer, val)
	}

	// flush (maybe)
	if len(self.buffer)/self.numCols > 100 {
		if err := self.Flush(); err != nil {
			return err
		}
	}

	return nil
}
