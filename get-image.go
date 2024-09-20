package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

	file, err := os.Open(files[0])
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	width, _ := strconv.Atoi(r.URL.Query().Get("width"))
	height, _ := strconv.Atoi(r.URL.Query().Get("height"))
	fmt.Printf("width: %d, height: %d\n", width, height)

	img, format, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Failed to decode image", http.StatusInternalServerError)
		return
	}

	if width > 0 || height > 0 {
		img = resize(img, width, height)
	}

	switch format {
	case "jpeg":
		err = jpeg.Encode(w, img, &jpeg.Options{Quality: 75})
		w.Header().Set("Content-Type", "image/jpeg")
	case "png":
		err = png.Encode(w, img)
		w.Header().Set("Content-Type", "image/png")
	default:
		http.Error(w, "Unsupported image format", http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, "Failed to encode image", http.StatusInternalServerError)
		return
	}
}

func resize(img image.Image, width, height int) image.Image {
	if width == 0 {
		width = img.Bounds().Dx() * height / img.Bounds().Dy()
	}
	if height == 0 {
		height = img.Bounds().Dy() * width / img.Bounds().Dx()
	}
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := x * img.Bounds().Dx() / width
			srcY := y * img.Bounds().Dy() / height
			newImg.Set(x, y, img.At(srcX, srcY))
		}
	}
	return newImg
}
