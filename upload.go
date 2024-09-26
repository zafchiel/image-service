package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

type UploadResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	ID      uint   `json:"id"`
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

		responses = append(responses, *response)
	}

	// Send response with all processed files
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)

}

func processUploadedFile(file *multipart.File, header *multipart.FileHeader) (*UploadResponse, error) {
	if err := validateImage(header); err != nil {
		return nil, err
	}

	fileBytes, err := io.ReadAll(*file)
	if err != nil {
		return nil, err
	}

	fileHash := generateFileHash(fileBytes)
	fileExt := filepath.Ext(header.Filename)
	newFilename := fileHash + fileExt

	var existingFile ImageMetadata

	// Check if file already exists
	result := db.Where("filename = ?", newFilename).First(&existingFile)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("Something went wrong")
	}
	if result.RowsAffected > 0 {
		return &UploadResponse{
			Success: true,
			ID:      existingFile.ID,
			Message: "File already exists",
			URL:     fmt.Sprintf("http://localhost:8080/image/%d", existingFile.ID),
		}, nil
	}

	// Save the file to the storage
	err = storage.Save(newFilename, bytes.NewReader(fileBytes))
	if err != nil {
		return nil, err
	}

	newFile := ImageMetadata{
		Filename: newFilename,
		Format:   fileExt,
		Size:     header.Size,
	}
	// Save file metadata to the database
	result = db.Create(&newFile)
	if result.Error != nil {
		return nil, result.Error
	}

	return &UploadResponse{
		Success: true,
		ID:      newFile.ID,
		Message: fmt.Sprintf("File %s uploaded successfully", header.Filename),
		URL:     fmt.Sprintf("http://localhost:8080/image/%d", newFile.ID),
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

func saveFile(fileBytes []byte, filename string) (string, error) {
	newID := filename[:8]
	destination, err := os.Create(filepath.Join("assets", filename))
	if err != nil {
		return "", err
	}
	defer destination.Close()

	_, err = destination.Write(fileBytes)
	return newID, err
}
