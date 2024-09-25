package main

import "io"

type Storage interface {
	Save(filename string, file io.Reader) error
	Get(filename string) (io.Reader, error)
	Delete(filename string) error
}

type LocalStorage struct {
	// Base path to store the files
	root string
}
