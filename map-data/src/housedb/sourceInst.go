package housedb

import (
	"context"
	"database/sql"
	"io"
	"time"
)

type SourceInst struct {
	Data io.ReadCloser

	ctx   context.Context
	db    *sql.Tx
	isNew bool
	name  string
	etag  *string
}

func (self *SourceInst) MakeRecord() error {
	now := time.Now()
	var err error
	if self.isNew {
		sql := "INSERT INTO source(name, etag, last_checked) VALUES ($1, $2, $3)"
		_, err = self.db.ExecContext(self.ctx, sql, self.name, self.etag, now.Unix())
	} else {
		sql := "UPDATE source SET etag = $1, last_checked = $2 WHERE name = $3"
		_, err = self.db.ExecContext(self.ctx, sql, self.etag, now.Unix(), self.name)
	}
	self.isNew = false
	return err
}
