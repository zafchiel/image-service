package main

import (
	"fmt"
	"image/jpeg"
	"image/png"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
)

type ImageFormat string

const (
	JPG  ImageFormat = "jpg"
	JPEG ImageFormat = "jpeg"
	PNG  ImageFormat = "png"
)

var supportedFormats = []ImageFormat{JPG, JPEG, PNG}

func getImage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if len(id) != 8 {
		http.Error(w, "Invalid image ID", http.StatusBadRequest)
		return
	}

	files, err := filepath.Glob("assets/" + id + "*")
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

	format := ImageFormat(r.URL.Query().Get("format"))
	if format == "" {
		format = ImageFormat(strings.TrimPrefix(filepath.Ext(files[0]), "."))
	}

	switch format {
	case JPG, JPEG:
		w.Header().Set("Content-Type", "image/jpeg")
		if err := jpeg.Encode(w, resized, &jpeg.Options{Quality: 75}); err != nil {
			http.Error(w, "Failed to encode image", http.StatusInternalServerError)
			return
		}
	case PNG:
		w.Header().Set("Content-Type", "image/png")
		if err := png.Encode(w, resized); err != nil {
			http.Error(w, "Failed to encode image", http.StatusInternalServerError)
			return
		}
	default:
		err := fmt.Sprintf("Unsupported image format: %s, use one of the following formats: %s", format, supportedFormats)
		http.Error(w, err, http.StatusBadRequest)
		return
	}
}
