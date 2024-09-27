package handlers

import (
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/url"
	"strconv"

	"github.com/anthonynsimon/bild/adjust"
	"github.com/anthonynsimon/bild/blur"
	"github.com/anthonynsimon/bild/effect"
	"github.com/anthonynsimon/bild/transform"
	"github.com/zafchiel/image-service/internal/errors"
	"github.com/zafchiel/image-service/internal/models"
)

type ImageFormat string

const (
	JPG  ImageFormat = "jpg"
	JPEG ImageFormat = "jpeg"
	PNG  ImageFormat = "png"
)

var supportedFormats = []ImageFormat{JPG, JPEG, PNG}

type GetImageHandler struct {
	app *App
}

func NewGetImageHandler(app *App) *GetImageHandler {
	return &GetImageHandler{app: app}
}

func (h *GetImageHandler) Handle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, errors.ErrInvalidID.Error(), http.StatusBadRequest)
		return
	}

	var imageMetadata models.ImageMetadata
	result := h.app.DB.First(&imageMetadata, id)
	if result.Error != nil {
		http.Error(w, errors.ErrImageNotFound.Error(), http.StatusNotFound)
		return
	}

	image, err := h.app.Storage.Get(imageMetadata.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	img, err := applyImageTransformations(image, r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/"+imageMetadata.Format)
	if imageMetadata.Format == "png" {
		png.Encode(w, img)
	} else {
		jpeg.Encode(w, img, nil)
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
