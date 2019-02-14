package storage

import (
	"errors"
	"log"
)

type StaticErr string

// implement io.Reader
func (sr StaticErr) Read(p []byte) (int, error) {
	log.Println(sr)
	return 0, errors.New(string(sr))
}

func (sr StaticErr) Close() error {
	return nil
}
