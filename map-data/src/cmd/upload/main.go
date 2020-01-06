package main

import (
	"encoding/json"
	"flag"
	"fmt"
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

	// get AWS creds from Mapbox
	mapbox := NewMapbox(mapboxToken, username)
	awsCreds, err := mapbox.MakeAwsCreds()
	if err != nil {
		return err
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
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	tilesetF, err := os.Open(tilesetPath)
	if err != nil {
		return err
	}
	defer tilesetF.Close()
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(awsCreds.Bucket),
		Key:    aws.String(awsCreds.Key),
		Body:   tilesetF,
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

	// get Mapbox token
	mapboxToken, err := accessToken()
	if err != nil {
		return err
	}

	// look for existing style
	log.Println("Looking for existing style")
	mapbox := NewMapbox(mapboxToken, username)
	styleInfos, err := mapbox.ListStyles()
	if err != nil {
		return err
	}
	var styleID *string
	for _, styleInfo := range styleInfos {
		if styleInfo.Name == styleName {
			styleID = &styleInfo.ID
			break
		}
	}

	// upload style
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	if styleID == nil {
		log.Println("Making new style")
		err = mapbox.CreateStyle(f)
	} else {
		log.Println("Updating existing style")
		err = mapbox.UpdateStyle(*styleID, f)
	}
	if err != nil {
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
