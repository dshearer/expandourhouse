package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

const gJSONContentType = "application/json"
const gContentTypeHeader = "Content-Type"

type Style struct {
	Version  int    `json:"version"`
	Name     string `json:"name"`
	ID       string `json:"id"`
	Owner    string `json:"owner"`
	Created  string `json:"created"`
	Modified string `json:"modified"`
}

type MapboxAwsCreds struct {
	AccessKeyID     string `json:"accessKeyId"`
	Bucket          string `json:"bucket"`
	Key             string `json:"key"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
	URL             string `json:"url"`
}

type errMsg struct {
	Message string `json:"message"`
}

func makeErrFromResp(resp *http.Response, msg string) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	defaultMsg := fmt.Sprintf("%v (HTTP %v)", msg, resp.StatusCode)
	if resp.Body == nil {
		return errors.New(defaultMsg)
	}

	decoder := json.NewDecoder(resp.Body)
	var contents errMsg
	if err := decoder.Decode(&contents); err != nil {
		return errors.New(defaultMsg)
	}
	return fmt.Errorf("%v: %v", msg, contents.Message)
}

type Mapbox struct {
	accessToken string
	username    string
	client      http.Client
}

func NewMapbox(accessToken, username string) Mapbox {
	if len(accessToken) == 0 {
		panic("Empty access token")
	}
	if len(username) == 0 {
		panic("Empty username")
	}
	return Mapbox{accessToken: accessToken, username: username}
}

func (mb *Mapbox) makeUrl(uriFormat string) string {
	uri := strings.ReplaceAll(uriFormat, "{user}", mb.username)
	return fmt.Sprintf("https://api.mapbox.com%s?access_token=%s", uri, mb.accessToken)
}

func (mb *Mapbox) ListStyles() ([]*Style, error) {
	// make request
	resp, err := mb.client.Get(mb.makeUrl("/styles/v1/{user}"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := makeErrFromResp(resp, "Couldn't list styles"); err != nil {
		return nil, err
	}

	// parse response
	var styles []*Style
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&styles); err != nil {
		return nil, err
	}
	return styles, nil
}

func (mb *Mapbox) CreateStyle(r io.Reader) error {
	// make request
	resp, err := mb.client.Post(mb.makeUrl("/styles/v1/{user}"),
		gJSONContentType, r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := makeErrFromResp(resp, "Couldn't make style"); err != nil {
		return err
	}
	return nil
}

func (mb *Mapbox) UpdateStyle(styleID string, r io.Reader) error {
	// make request
	req, err := http.NewRequest(http.MethodPatch, mb.makeUrl("/styles/v1/{user}/"+styleID), r)
	if err != nil {
		return err
	}
	req.Header.Add(gContentTypeHeader, gJSONContentType)
	resp, err := mb.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := makeErrFromResp(resp, "Couldn't update style"); err != nil {
		return err
	}
	return nil
}

func (mb *Mapbox) DeleteStylesWithName(name string) error {
	// list styles
	styles, err := mb.ListStyles()
	if err != nil {
		return err
	}

	// filter by name
	var stylesWithName []*Style
	for _, style := range styles {
		if style.Name == name {
			stylesWithName = append(stylesWithName, style)
		}
	}

	// get etags
	etags := make(map[string]string)
	for _, style := range stylesWithName {
		// send request
		req, err := http.NewRequest(http.MethodGet, mb.makeUrl("/styles/v1/{user}/"+style.ID), nil)
		if err != nil {
			return err
		}
		resp, err := mb.client.Do(req)
		if err != nil {
			return err
		}

		// get etag
		defer resp.Body.Close()
		if resp.StatusCode == 404 {
			continue
		}
		if err := makeErrFromResp(resp, "Failed to get style"); err != nil {
			return err
		}
		etag := resp.Header.Get("ETag")
		if len(etag) == 0 {
			return errors.New("No etag for style")
		}

		// check name
		var newStyle Style
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&newStyle); err != nil {
			return err
		}
		if newStyle.Name != name {
			continue
		}

		etags[style.ID] = etag
	}

	// delete styles
	nbrDeleted := 0
	for id, etag := range etags {
		// send request
		req, err := http.NewRequest(http.MethodDelete, mb.makeUrl("/styles/v1/{user}/"+id), nil)
		if err != nil {
			return err
		}
		req.Header.Add("If-Match", etag)
		resp, err := mb.client.Do(req)
		if err != nil {
			return err
		}

		// handle response
		if resp.StatusCode == 412 {
			continue
		}
		if resp.StatusCode == 404 {
			/* This sometimes happens. There's something wrong with the API... */
			continue
		}
		if err := makeErrFromResp(resp, fmt.Sprintf("Failed to delete style %v", id)); err != nil {
			return err
		}
		nbrDeleted++
	}

	log.Printf("Deleted %v styles\n", nbrDeleted)
	return nil
}

func (mb *Mapbox) CreateUpload(tilesetID, tilesetName, url string) error {
	// make request body
	body := map[string]string{
		"name":    tilesetName,
		"tileset": tilesetID,
		"url":     url,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// make request
	resp, err := mb.client.Post(mb.makeUrl("/uploads/v1/{user}"),
		gJSONContentType, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := makeErrFromResp(resp, "Couldn't create upload"); err != nil {
		return err
	}
	return nil
}

func (mb *Mapbox) MakeAwsCreds() (*MapboxAwsCreds, error) {
	// make request
	resp, err := mb.client.Post(mb.makeUrl("/uploads/v1/{user}/credentials"),
		"", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := makeErrFromResp(resp, "Couldn't make AWS creds"); err != nil {
		return nil, err
	}

	// parse response
	var creds MapboxAwsCreds
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

func (mb *Mapbox) TilesetExists(id string) (bool, error) {
	// make request
	resp, err := mb.client.Get(mb.makeUrl("/tilesets/v1/{user}"))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if err := makeErrFromResp(resp, "Couldn't get tileset "+id); err != nil {
		return false, err
	}

	// handle response
	var tilesets []map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&tilesets); err != nil {
		return false, errors.Wrapf(err, "Failed to parse tileset list")
	}
	for _, ts := range tilesets {
		if ts["id"].(string) == id {
			return true, nil
		}
	}
	return false, nil
}
