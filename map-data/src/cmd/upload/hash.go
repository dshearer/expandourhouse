package main

import (
	"crypto/md5"
	"encoding/binary"
	"io"
	"io/ioutil"
	"os"
	"path"
)

func HashFilePath(tilesetPath string) string {
	hashFileName := "." + path.Base(tilesetPath) + ".hash"
	return path.Join(path.Dir(tilesetPath), hashFileName)
}

func hashTileset(statesOrDistricts, stylePath, tilesetPath string) ([]byte, error) {
	hash := md5.New()

	hashLength := func(length int64) error {
		return binary.Write(hash, binary.BigEndian, length)
	}
	hashString := func(s string) error {
		if err := hashLength(int64(len(s))); err != nil {
			return err
		}
		_, err := io.WriteString(hash, s)
		return err
	}
	hashFile := func(path string) error {
		stat, err := os.Stat(stylePath)
		if err != nil {
			return err
		}
		if err := hashLength(stat.Size()); err != nil {
			return err
		}
		f, err := os.Open(stylePath)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(hash, f)
		return err
	}

	if err := hashString(statesOrDistricts); err != nil {
		return nil, err
	}
	if err := hashFile(stylePath); err != nil {
		return nil, err
	}
	if err := hashFile(tilesetPath); err != nil {
		return nil, err
	}
	h := hash.Sum(nil)
	return h[:], nil
}

func ShouldUploadTileset(statesOrDistricts, stylePath, tilesetPath string) (bool, error) {
	// get old hash
	f, err := os.Open(HashFilePath(tilesetPath))
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	defer f.Close()
	oldHash, err := ioutil.ReadAll(f)
	if err != nil {
		return false, err
	}

	// get new hash
	newHash, err := hashTileset(statesOrDistricts, stylePath, tilesetPath)
	if err != nil {
		return false, err
	}

	// compare hashes
	if len(oldHash) != len(newHash) {
		return true, nil
	}
	for i := 0; i < len(oldHash); i++ {
		if oldHash[i] != newHash[i] {
			return true, nil
		}
	}
	return false, nil
}

func RecordUploadedTileset(statesOrDistricts, stylePath, tilesetPath string) error {
	// get new hash
	newHash, err := hashTileset(statesOrDistricts, stylePath, tilesetPath)
	if err != nil {
		return err
	}

	// save it
	f, err := os.Create(HashFilePath(tilesetPath))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(newHash)
	return err
}
