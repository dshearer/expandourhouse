package sourceinst

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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

func FetchHttpSourceIfChanged(ctx context.Context, name, url string, db *sql.Tx) (*SourceInst, error) {
	// look up source
	rows, err := db.QueryContext(ctx, "SELECT etag, last_checked FROM source WHERE name = $1", name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var oldEtag *string
	var lastChecked *int
	isNew := false
	if !rows.Next() {
		isNew = true
	} else if err := rows.Scan(&oldEtag, &lastChecked); err != nil {
		return nil, err
	}

	// have we checked recently?
	if !isNew && time.Now().Before(time.Unix(int64(*lastChecked), 0).Add(24*time.Hour)) {
		/* assume not modified */
		return nil, nil
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

func FetchLocalSourceIfChanged(ctx context.Context, name string, data io.ReadSeeker, db *sql.Tx) (*SourceInst, error) {
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

	// compute hash
	h := md5.New()
	if _, err := data.Seek(0, 0); err != nil {
		return nil, err
	}
	if _, err := io.Copy(h, data); err != nil {
		return nil, err
	}
	hash := base64.RawStdEncoding.EncodeToString(h.Sum(nil))

	// compare hashes
	if !isNew && hash == *oldEtag {
		/* not modified */
		return nil, nil
	}

	// make new inst
	if _, err := data.Seek(0, 0); err != nil {
		return nil, err
	}
	return &SourceInst{
		Data:  ioutil.NopCloser(data),
		ctx:   ctx,
		db:    db,
		isNew: isNew,
		name:  name,
		etag:  &hash,
	}, nil
}
