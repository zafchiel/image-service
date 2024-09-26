package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/anthonynsimon/bild/adjust"
	"github.com/anthonynsimon/bild/blur"
	"github.com/anthonynsimon/bild/effect"
	"github.com/anthonynsimon/bild/transform"
)

func getImage(w http.ResponseWriter, r *http.Request) {
	// Validate image ID
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	var imageMetadata ImageMetadata
	// Search for the image metadata in the database
	result := db.First(&imageMetadata, id)
	if result.Error != nil {
		http.Error(w, fmt.Sprintf("Image of ID %v not found", id), http.StatusNotFound)
		return
	}

	// Find and open the image image
	image, err := storage.Get(imageMetadata.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Apply image transformations
	img, err := applyImageTransformations(image, r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine output format and encode image
	if err := encodeAndSendImage(w, img, r.URL.Query().Get("format"), imageMetadata.Format); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func applyImageTransformations(img image.Image, query url.Values) (image.Image, error) {
	width, _ := strconv.Atoi(query.Get("width"))
	height, _ := strconv.Atoi(query.Get("height"))
	resized := img

	if width == 0 || height == 0 {
		resized = img
	} else {
		resized = transform.Resize(img, width, height, transform.Linear)
	}

	// Apply other transformations
	if blurRadius, err := strconv.ParseFloat(query.Get("blur"), 64); err == nil && blurRadius > 0 {
		resized = blur.Gaussian(resized, blurRadius)
	}

	if brightness, err := strconv.ParseFloat(query.Get("brightness"), 64); err == nil {
		resized = adjust.Brightness(resized, brightness)
	}

	if contrast, err := strconv.ParseFloat(query.Get("contrast"), 64); err == nil {
		resized = adjust.Contrast(resized, contrast)
	}

	if query.Get("grayscale") == "true" {
		resized = effect.Grayscale(resized)
	}

	if query.Get("sepia") == "true" {
		resized = effect.Sepia(resized)
	}

	if query.Get("invert") == "true" {
		resized = effect.Invert(resized)
	}

	if rotation, err := strconv.ParseFloat(query.Get("rotate"), 64); err == nil {
		resized = transform.Rotate(resized, rotation, nil)
	}

	if query.Get("fliph") == "true" {
		resized = transform.FlipH(resized)
	}

	if query.Get("flipv") == "true" {
		resized = transform.FlipV(resized)
	}

	return resized, nil
}

func encodeAndSendImage(w http.ResponseWriter, img image.Image, formatParam, defaultExt string) error {
	format := ImageFormat(formatParam)
	if format == "" {
		format = ImageFormat(strings.TrimPrefix(defaultExt, "."))
	}

	switch format {
	case JPG, JPEG:
		w.Header().Set("Content-Type", "image/jpeg")
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 75})
	case PNG:
		w.Header().Set("Content-Type", "image/png")
		return png.Encode(w, img)
	default:
		return fmt.Errorf("unsupported image format: %s, use one of the following formats: %v", format, supportedFormats)
	}
}
