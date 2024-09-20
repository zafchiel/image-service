package main

import (
	"crypto/sha256"
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
	router.HandleFunc("GET /", hello)
	router.HandleFunc("POST /upload", uploadImage)
	router.HandleFunc("GET /image/{id}", getImage)

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

	fileExt := filepath.Ext(header.Filename)

	// Read file into memory
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Generate a SHA-256 hash of the file
	hash := sha256.Sum256(fileBytes)
	fileHash := hex.EncodeToString(hash[:])

	// Check if a file with this hash already exists
	existingFilePath := filepath.Join("images", fileHash+fileExt)
	if _, err := os.Stat(existingFilePath); err == nil {
		// File already exists, return its information
		res := struct {
			ID      string `json:"id"`
			Message string `json:"message"`
			URL     string `json:"url"`
		}{
			ID:      fileHash[:8], // Use first 8 characters of hash as ID
			Message: "File already exists",
			URL:     fmt.Sprintf("http://localhost:8080/image/%s", fileHash[:8]),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	newFilename := fileHash + fileExt
	destination, err := os.Create("images/" + newFilename)
	if err != nil {
		http.Error(w, "Failed to create file on server", http.StatusInternalServerError)
		return
	}
	defer destination.Close()

	_, err = destination.Write(fileBytes)
	if err != nil {
		http.Error(w, "Failed to save file on server", http.StatusInternalServerError)
		return
	}

	res := struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		URL     string `json:"url"`
	}{
		ID:      fileHash[:8],
		Message: fmt.Sprintf("File %s uploaded successfully", header.Filename),
		URL:     fmt.Sprintf("http://localhost:8080/image/%s", fileHash[:8]),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

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

	filename := filepath.Base(files[0])
	file, err := os.Open(files[0])
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var contentType string
	switch filepath.Ext(filename) {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	default:
		http.Error(w, "Unsupported image format", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Failed to copy file to response writer", http.StatusInternalServerError)
		return
	}
}
