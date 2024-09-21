package main

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
)

func getImage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if len(id) != 8 {
		http.Error(w, "Invalid image ID", http.StatusBadRequest)
		return
	}

	files, err := filepath.Glob("images/" + id + "*")
	if err != nil {
		http.Error(w, "Failed to find file", http.StatusInternalServerError)
		return
	}

	if len(files) == 0 {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	if len(files) > 1 {
		http.Error(w, "Multiple images found", http.StatusInternalServerError)
		return
	}

	file, err := imgio.Open(files[0])
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}

	width, _ := strconv.Atoi(r.URL.Query().Get("width"))
	height, _ := strconv.Atoi(r.URL.Query().Get("height"))

	resized := transform.Resize(file, width, height, transform.Linear)

	fileType := filepath.Ext(files[0])
	var encoder imgio.Encoder
	switch fileType {
	case ".jpg", ".jpeg":
		encoder = imgio.JPEGEncoder(75)
	case ".png":
		encoder = imgio.PNGEncoder()
	default:
		http.Error(w, "Unsupported image format", http.StatusInternalServerError)
		return
	}

	encoder(w, resized)
}
