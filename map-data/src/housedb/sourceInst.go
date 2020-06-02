package housedb

import (
	"context"
	"database/sql"
	"io"
)

type SourceInst struct {
	Data io.ReadCloser

	ctx   context.Context
	db    *sql.DB
	isNew bool
	name  string
	etag  *string
}

func (self *SourceInst) MakeRecord() error {
	var err error
	if self.isNew {
		sql := "INSERT INTO source(name, etag) VALUES ($1, $2)"
		_, err = self.db.ExecContext(self.ctx, sql, self.name, self.etag)
	} else {
		sql := "UPDATE source SET etag = $1 WHERE name = $2"
		_, err = self.db.ExecContext(self.ctx, sql, self.etag, self.name)
	}
	self.isNew = false
	return err
}
