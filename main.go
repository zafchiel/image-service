package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/zafchiel/image-service/middleware"
)

const (
	port          = ":8080"
	maxUploadSize = 10 << 20 // 10 MB
)

func main() {
	router := http.NewServeMux()
	router.HandleFunc("/", hello)
	router.HandleFunc("POST /upload", uploadImage)

	router.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))

	server := http.Server{
		Addr:    port,
		Handler: middleware.Logger(router),
	}

	fmt.Println("Server is running on port", port)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!")
}

func generateID(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form data
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "The uploaded file is too big. Please upload a file up to 10MB", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "No image file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		http.Error(w, "Only jpeg and png images are allowed", http.StatusBadRequest)
		return
	}

	if header.Size > maxUploadSize {
		http.Error(
			w,
			fmt.Sprintf("The uploaded image is too big: %v. Please upload an image up to %v", header.Size, maxUploadSize),
			http.StatusBadRequest,
		)
		return
	}

	id := generateID(4) // Generate an 8-byte (16 character) ID
	fileExt := filepath.Ext(header.Filename)
	newFilename := id + fileExt

	destination, err := os.Create("images/" + newFilename)
	if err != nil {
		http.Error(w, "Failed to create file on server", http.StatusInternalServerError)
		return
	}
	defer destination.Close()

	_, err = io.Copy(destination, file)
	if err != nil {
		http.Error(w, "Failed to save file on server", http.StatusInternalServerError)
		return
	}

	res := struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		URL     string `json:"url"`
	}{
		ID:      id,
		Message: fmt.Sprintf("File %s uploaded successfully", header.Filename),
		URL:     fmt.Sprintf("http://localhost:8080/images/%s", newFilename),
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
}
