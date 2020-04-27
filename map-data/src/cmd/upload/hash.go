package main

import (
	"crypto/md5"
	"io"
	"io/ioutil"
	"os"
	"path"
)

func HashFilePath(filePath string) string {
	hashFileName := "." + path.Base(filePath) + ".hash"
	return path.Join(path.Dir(filePath), hashFileName)
}

func hashFile(filePath string) ([]byte, error) {
	hash := md5.New()
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	_, err = io.Copy(hash, f)
	if err != nil {
		return nil, err
	}
	h := hash.Sum(nil)
	return h[:], nil
}

func FileHasChanged(filePath string) (bool, error) {
	// get old hash
	f, err := os.Open(HashFilePath(filePath))
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
	newHash, err := hashFile(filePath)
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

func RecordUploaded(filePath string) error {
	// get new hash
	newHash, err := hashFile(filePath)
	if err != nil {
		return err
	}

	// save it
	f, err := os.Create(HashFilePath(filePath))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(newHash)
	return err
}
