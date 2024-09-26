package main

import (
	"image"
	"io"
	"os"
	"path/filepath"

	"github.com/anthonynsimon/bild/imgio"
)

type Storage interface {
	Save(filename string, content io.Reader) error
	Get(filename string) (image.Image, error)
	Delete(filename string) error
}

type LocalStorage struct {
	// Base path to store the files
	root string
}

func (ls *LocalStorage) Save(filename string, content io.Reader) error {
	fullPath := filepath.Join(ls.root, filename)
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, content)
	return err
}

func (ls *LocalStorage) Get(filename string) (image.Image, error) {
	fullPath := filepath.Join(ls.root, filename)
	return imgio.Open(fullPath)
}

func (ls *LocalStorage) Delete(filename string) error {
	fullPath := filepath.Join(ls.root, filename)
	return os.Remove(fullPath)
}
