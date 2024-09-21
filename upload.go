package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type UploadResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	URL     string `json:"url"`
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "No image file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err := validateImage(header); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	fileHash := generateFileHash(fileBytes)
	fileExt := filepath.Ext(header.Filename)
	newFilename := fileHash + fileExt

	if existingFile(fileHash, fileExt) {
		sendResponse(w, UploadResponse{
			ID:      fileHash[:8],
			Message: "File already exists",
			URL:     fmt.Sprintf("http://localhost:8080/image/%s", fileHash[:8]),
		}, http.StatusOK)
		return
	}

	if err := saveFile(fileBytes, newFilename); err != nil {
		http.Error(w, "Failed to save file on server", http.StatusInternalServerError)
		return
	}

	sendResponse(w, UploadResponse{
		ID:      fileHash[:8],
		Message: fmt.Sprintf("File %s uploaded successfully", header.Filename),
		URL:     fmt.Sprintf("http://localhost:8080/image/%s", fileHash[:8]),
	}, http.StatusCreated)
}

func validateImage(header *multipart.FileHeader) error {
	if header.Size > maxUploadSize {
		return fmt.Errorf("the uploaded image is too big: %v. Please upload an image up to %v", header.Size, maxUploadSize)
	}

	contentType := ImageFormat(strings.Split(header.Header.Get("Content-Type"), "/")[1])
	var format ImageFormat
	for _, f := range supportedFormats {
		if f == contentType {
			format = f
			break
		}
	}

	if format == "" {
		return fmt.Errorf("unsupported image format: %s, upload one of the following: %v", contentType, supportedFormats)
	}

	return nil
}

func generateFileHash(fileBytes []byte) string {
	hash := sha256.Sum256(fileBytes)
	return hex.EncodeToString(hash[:])
}

func existingFile(fileHash, fileExt string) bool {
	_, err := os.Stat(filepath.Join("assets", fileHash+fileExt))
	return err == nil
}

func saveFile(fileBytes []byte, filename string) error {
	destination, err := os.Create(filepath.Join("assets", filename))
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = destination.Write(fileBytes)
	return err
}

func sendResponse(w http.ResponseWriter, response UploadResponse, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
