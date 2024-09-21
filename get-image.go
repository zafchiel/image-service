package main

import (
	"fmt"
	"image/jpeg"
	"image/png"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/anthonynsimon/bild/adjust"
	"github.com/anthonynsimon/bild/blur"
	"github.com/anthonynsimon/bild/effect"
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

	// Apply effects based on query parameters
	width, _ := strconv.Atoi(r.URL.Query().Get("width"))
	height, _ := strconv.Atoi(r.URL.Query().Get("height"))

	resized := transform.Resize(file, width, height, transform.Linear)

	if blurRadius, err := strconv.ParseFloat(r.URL.Query().Get("blur"), 64); err == nil && blurRadius > 0 {
		resized = blur.Gaussian(resized, blurRadius)
	}

	if brightness, err := strconv.ParseFloat(r.URL.Query().Get("brightness"), 64); err == nil {
		resized = adjust.Brightness(resized, brightness)
	}

	if contrast, err := strconv.ParseFloat(r.URL.Query().Get("contrast"), 64); err == nil {
		resized = adjust.Contrast(resized, contrast)
	}

	if r.URL.Query().Get("grayscale") == "true" {
		resized = effect.Grayscale(resized)
	}

	if r.URL.Query().Get("sepia") == "true" {
		resized = effect.Sepia(resized)
	}

	if r.URL.Query().Get("invert") == "true" {
		resized = effect.Invert(resized)
	}

	if rotation, err := strconv.ParseFloat(r.URL.Query().Get("rotate"), 64); err == nil {
		resized = transform.Rotate(resized, rotation, nil)
	}

	if r.URL.Query().Get("fliph") == "true" {
		resized = transform.FlipH(resized)
	}

	if r.URL.Query().Get("flipv") == "true" {
		resized = transform.FlipV(resized)
	}

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
