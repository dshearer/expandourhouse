package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
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

func FetchSourceIfChanged(ctx context.Context, name, url string, db *sql.DB) (*SourceInst, error) {
	// look up source
	rows, err := db.QueryContext(ctx, "SELECT etag FROM source WHERE name = $1", name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var oldEtag *string
	isNew := false
	if !rows.Next() {
		isNew = true
	} else if err := rows.Scan(&oldEtag); err != nil {
		return nil, err
	}

	// get source
	var client http.Client
	var resp *http.Response
	if isNew || oldEtag == nil {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode > 299 {
			resp.Body.Close()
			return nil, fmt.Errorf("Got HTTP status: %v", resp.Status)
		}

	} else {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("If-None-Match", *oldEtag)
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode > 299 && resp.StatusCode != 304 {
			resp.Body.Close()
			return nil, fmt.Errorf("Got HTTP status: %v", resp.Status)
		}
		if resp.StatusCode == 304 {
			/* not modified */
			return nil, nil
		}
	}

	etag := resp.Header.Get("ETag")
	etagP := &etag
	if len(etag) == 0 {
		etagP = nil
	}
	return &SourceInst{
		Data:  resp.Body,
		ctx:   ctx,
		db:    db,
		isNew: isNew,
		name:  name,
		etag:  etagP,
	}, nil
}
