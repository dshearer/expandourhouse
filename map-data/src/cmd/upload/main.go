package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	stylemetadata "expandourhouse.com/mapdata/styleMetadata"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

const gUsage = "usage: upload (style MAPBOX_USER STYLE_FILE | " +
	"states MAPBOX_USER STYLE_FILE STATES_TILESET | " +
	"districts MAPBOX_USER STYLE_FILE DISTRICTS_TILESET)\n"

var errInvalidArgs = errors.New("Invalid args")

const gTokenEnvVar = "MAPBOX_WRITE_SCOPE_ACCESS_TOKEN"

type reportingReader struct {
	r        io.Reader
	callback func(int)
}

func (self *reportingReader) Read(p []byte) (int, error) {
	n, err := self.r.Read(p)
	if n > 0 {
		self.callback(n)
	}
	return n, err
}

func accessToken() (string, error) {
	tok, ok := os.LookupEnv(gTokenEnvVar)
	if !ok {
		return "", fmt.Errorf("Missing env var: %s", gTokenEnvVar)
	}
	return tok, nil
}

func uploadTileset(statesOrDistricts, stylePath, tilesetPath, username,
	mapboxToken string) error {

	// get some stuff from the style
	f, err := os.Open(stylePath)
	if err != nil {
		return err
	}
	defer f.Close()
	var style map[string]interface{}
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&style); err != nil {
		return err
	}
	metadata, err := stylemetadata.Get(style, username)
	if err != nil {
		return err
	}
	var tilesetID, tilesetName string
	if statesOrDistricts == "states" {
		tilesetID = metadata.StatesTilesetID
		tilesetName = metadata.StatesTilesetName
	} else {
		tilesetID = metadata.DistrictsTilesetID
		tilesetName = metadata.DistrictsTilesetName
	}

	// check if tileset exists on Mapbox
	mapbox := NewMapbox(mapboxToken, username)
	exists, err := mapbox.TilesetExists(tilesetID)
	if err != nil {
		return err
	}

	if exists {
		// check if file has changed
		hasChanged, err := FileHasChanged(tilesetPath)
		if err != nil {
			return err
		}
		if !hasChanged {
			fmt.Println("No need to upload: tileset hasn't changed")
			return nil
		}
	}

	// get AWS creds from Mapbox
	awsCreds, err := mapbox.MakeAwsCreds()
	if err != nil {
		return err
	}

	// get tileset file size
	stat, err := os.Stat(tilesetPath)
	if err != nil {
		return err
	}

	// make tileset reader
	progBar := ProgBar{Total: int(stat.Size())}
	tilesetF, err := os.Open(tilesetPath)
	if err != nil {
		return err
	}
	defer tilesetF.Close()
	tilesetF2 := reportingReader{
		r: tilesetF,
		callback: func(n int) {
			progBar.AddProgress(n)
		},
	}

	// stage tileset on AWS S3
	fmt.Print("Staging tileset on AWS...")
	creds := credentials.NewStaticCredentials(awsCreds.AccessKeyID,
		awsCreds.SecretAccessKey, awsCreds.SessionToken)
	awsSess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: creds,
	})
	if err != nil {
		return err
	}
	uploader := s3manager.NewUploader(awsSess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(awsCreds.Bucket),
		Key:    aws.String(awsCreds.Key),
		Body:   &tilesetF2,
	})
	if err != nil {
		return err
	}
	fmt.Println("done.")

	// send it to Mapbox
	err = mapbox.CreateUpload(tilesetID, tilesetName, awsCreds.URL)
	if err != nil {
		return err
	}

	// save hash
	if err := RecordUploaded(tilesetPath); err != nil {
		return err
	}

	return nil
}

func uploadStyle(args []string) error {
	/* usage: style MAPBOX_USER STYLE_FILE */

	// get args
	if len(args) != 2 {
		return errInvalidArgs
	}
	username := args[0]
	stylePath := args[1]

	// get Mapbox token
	mapboxToken, err := accessToken()
	if err != nil {
		return err
	}

	// get style name from style file
	f, err := os.Open(stylePath)
	if err != nil {
		return err
	}
	defer f.Close()
	var style map[string]interface{}
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&style); err != nil {
		return err
	}
	styleName := style["name"].(string)

	// delete existing styles
	mapbox := NewMapbox(mapboxToken, username)
	if err := mapbox.DeleteStylesWithName(styleName); err != nil {
		return err
	}

	// upload style
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	log.Printf("Making new style: %v\n", styleName)
	if err := mapbox.CreateStyle(f); err != nil {
		return err
	}

	return nil
}

func uploadStates(args []string) error {
	/* Usage: states MAPBOX_USER STYLE_FILE STATES_TILESET */

	// get args
	if len(args) != 3 {
		return errInvalidArgs
	}
	username := args[0]
	stylePath := args[1]
	tilesetPath := args[2]

	// get Mapbox token
	mapboxToken, err := accessToken()
	if err != nil {
		return err
	}

	return uploadTileset("states", stylePath, tilesetPath, username, mapboxToken)
}

func uploadDistricts(args []string) error {
	/* Usage: tiles MAPBOX_USER STYLE_FILE STATES_TILESET */

	// get args
	if len(args) != 3 {
		return errInvalidArgs
	}
	username := args[0]
	stylePath := args[1]
	tilesetPath := args[2]

	// get Mapbox token
	mapboxToken, err := accessToken()
	if err != nil {
		return err
	}

	return uploadTileset("tiles", stylePath, tilesetPath, username, mapboxToken)
}

func main() {
	log.SetOutput(os.Stderr)

	// get args
	flag.Parse()
	if flag.NArg() == 0 {
		os.Stderr.WriteString(gUsage)
		os.Exit(1)
	}
	cmd := flag.Arg(0)
	var f func([]string) error
	switch cmd {
	case "style":
		f = uploadStyle
	case "states":
		f = uploadStates
	case "districts":
		f = uploadDistricts
	default:
		os.Stderr.WriteString(gUsage)
		os.Exit(1)
	}

	// do command
	err := f(flag.Args()[1:])
	if err == errInvalidArgs {
		os.Stderr.WriteString(gUsage)
		os.Exit(1)
	} else if err != nil {
		log.Panic(err)
	}
}
