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
	Success bool   `json:"success"`
	Error   string `json:"error"`
	ID      string `json:"id"`
	Message string `json:"message"`
	URL     string `json:"url"`
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form data
	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Get the file headers for all uploaded files
	files := r.MultipartForm.File["image"]
	if len(files) == 0 {
		http.Error(w, "No image files uploaded", http.StatusBadRequest)
		return
	}

	// Process each uploaded file
	var responses []UploadResponse
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to open file: %v", err), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Process the file (validate, save, etc.)
		response, err := processUploadedFile(&file, fileHeader)
		if err != nil {
			responses = append(responses, UploadResponse{Success: false, Error: err.Error()})
			continue
		}

		responses = append(responses, response)
	}

	// Send response with all processed files
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)

}

func processUploadedFile(file *multipart.File, header *multipart.FileHeader) (UploadResponse, error) {

	if err := validateImage(header); err != nil {
		return UploadResponse{}, err
	}

	fileBytes, err := io.ReadAll(*file)
	if err != nil {
		return UploadResponse{}, err
	}

	fileHash := generateFileHash(fileBytes)
	fileExt := filepath.Ext(header.Filename)
	newFilename := fileHash + fileExt

	if existingFile(fileHash, fileExt) {
		return UploadResponse{
			Success: true,
			ID:      fileHash[:8],
			Message: "File already exists",
			URL:     fmt.Sprintf("http://localhost:8080/image/%s", fileHash[:8]),
		}, nil
	}

	if err := saveFile(fileBytes, newFilename); err != nil {
		return UploadResponse{}, err
	}

	return UploadResponse{
		Success: true,
		ID:      fileHash[:8],
		Message: fmt.Sprintf("File %s uploaded successfully", header.Filename),
		URL:     fmt.Sprintf("http://localhost:8080/image/%s", fileHash[:8]),
	}, nil
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
