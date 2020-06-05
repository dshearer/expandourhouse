package housedb

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"io"
	"io/ioutil"
)

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
