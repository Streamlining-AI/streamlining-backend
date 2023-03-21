package helper

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func CreateAndGetDir(path string) (string, error) {
	// get the full path of the directory
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	// create the directory if it does not exist
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		if err := os.MkdirAll(absPath, os.ModePerm); err != nil {
			return "", err
		}
	}

	// return the full path of the directory
	return absPath, nil
}

func ReadFile(filepath string) ([]byte, error) {
	// read the file contents
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return data, nil
}
