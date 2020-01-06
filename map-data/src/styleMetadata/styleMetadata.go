package stylemetadata

import (
	"errors"
	"fmt"
)

func makeStatesTilesetIDProp(username string) string {
	return fmt.Sprintf("%v:states-tileset-id", username)
}

func makeStatesTilesetNameProp(username string) string {
	return fmt.Sprintf("%v:states-tileset-name", username)
}

func makeDistrictsTilesetIDProp(username string) string {
	return fmt.Sprintf("%v:districts-tileset-id", username)
}

func makeDistrictsTilesetNameProp(username string) string {
	return fmt.Sprintf("%v:districts-tileset-name", username)
}

type StyleMetadata struct {
	StatesTilesetID      string
	StatesTilesetName    string
	DistrictsTilesetID   string
	DistrictsTilesetName string
}

func tryString(metadata map[string]interface{}, prop string) (string, bool) {
	val, ok := metadata[prop]
	if !ok {
		return "", false
	}
	s, ok := val.(string)
	if !ok {
		return "", false
	}
	return s, true
}

func Get(style map[string]interface{}, username string) (*StyleMetadata, error) {
	metadataObj, ok := style["metadata"].(map[string]interface{})
	if !ok {
		return nil, errors.New("Style doesn't have metadata")
	}
	var metadata StyleMetadata
	metadata.StatesTilesetID, ok = tryString(metadataObj, makeStatesTilesetIDProp(username))
	if !ok {
		return nil, errors.New("Metadata is missing states tileset ID")
	}
	metadata.StatesTilesetName, ok = tryString(metadataObj, makeStatesTilesetNameProp(username))
	if !ok {
		return nil, errors.New("Metadata is missing states tileset name")
	}
	metadata.DistrictsTilesetID, ok = tryString(metadataObj, makeDistrictsTilesetIDProp(username))
	if !ok {
		return nil, errors.New("Metadata is missing districts tileset ID")
	}
	metadata.DistrictsTilesetName, ok = tryString(metadataObj, makeDistrictsTilesetNameProp(username))
	if !ok {
		return nil, errors.New("Metadata is missing districts tileset name")
	}
	return &metadata, nil
}

func Set(style map[string]interface{}, username string, metadata *StyleMetadata) {
	metadataObj, ok := style["metadata"].(map[string]interface{})
	if !ok {
		metadataObj = make(map[string]interface{})
		style["metadata"] = metadataObj
	}
	metadataObj[makeStatesTilesetIDProp(username)] = metadata.StatesTilesetID
	metadataObj[makeStatesTilesetNameProp(username)] = metadata.StatesTilesetName
	metadataObj[makeDistrictsTilesetIDProp(username)] = metadata.DistrictsTilesetID
	metadataObj[makeDistrictsTilesetNameProp(username)] = metadata.DistrictsTilesetName
}
