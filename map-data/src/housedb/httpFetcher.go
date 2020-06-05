package housedb

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

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
