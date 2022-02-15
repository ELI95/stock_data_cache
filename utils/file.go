package utils

import (
	"encoding/gob"
	"os"
)

func Save(path string, object interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	return gob.NewEncoder(file).Encode(object)
}

func Load(path string, object interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	return gob.NewDecoder(file).Decode(object)
}
