package handlers

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
	"path/filepath"
	"strings"

	"github.com/zafchiel/image-service/internal/models"
	"gorm.io/gorm"
)

type UploadResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	ID      uint   `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
	URL     string `json:"url,omitempty"`
}

type UploadHandler struct {
	app *App
}

func NewUploadHandler(app *App) *UploadHandler {
	return &UploadHandler{app: app}
}

func (h *UploadHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if err := h.validateRequest(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["image"]
	responses := h.processFiles(files)

	h.sendResponse(w, responses)
}

func (h *UploadHandler) validateRequest(r *http.Request) error {
	r.Body = http.MaxBytesReader(nil, r.Body, h.app.Config.MaxUploadSize)

	if err := r.ParseMultipartForm(h.app.Config.MaxUploadSize); err != nil {
		return fmt.Errorf("failed to parse multipart form: %w", err)
	}

	if len(r.MultipartForm.File["image"]) == 0 {
		return errors.New("no image uploaded")
	}

	return nil
}

func (h *UploadHandler) processFiles(files []*multipart.FileHeader) []UploadResponse {
	responses := make([]UploadResponse, 0, len(files))
	for _, fileHeader := range files {
		response := h.processFile(fileHeader)
		responses = append(responses, response)
	}
	return responses
}

func (h *UploadHandler) processFile(fileHeader *multipart.FileHeader) UploadResponse {
	file, err := fileHeader.Open()
	if err != nil {
		return UploadResponse{Success: false, Error: fmt.Sprintf("Failed to open file: %v", err)}
	}
	defer file.Close()

	response, err := processUploadedFile(file, fileHeader, h.app)
	if err != nil {
		return UploadResponse{Success: false, Error: err.Error()}
	}

	return *response
}

func (h *UploadHandler) sendResponse(w http.ResponseWriter, responses []UploadResponse) {
	w.Header().Set("Content-Type", "application/json")
	statusCode := h.determineStatusCode(responses)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(responses)
}

func (h *UploadHandler) determineStatusCode(responses []UploadResponse) int {
	for _, response := range responses {
		if !response.Success {
			return http.StatusBadRequest
		}
	}
	return http.StatusOK
}

func processUploadedFile(file multipart.File, header *multipart.FileHeader, app *App) (*UploadResponse, error) {
	if err := validateImage(header, app.Config.MaxUploadSize); err != nil {
		return &UploadResponse{Success: false, Error: err.Error()}, nil
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return &UploadResponse{Success: false, Error: "Failed to read file"}, nil
	}

	fileHash := generateFileHash(fileBytes)
	fileExt := filepath.Ext(header.Filename)
	newFilename := fileHash + fileExt

	existingFile, err := checkExistingFile(app.DB, newFilename)
	if err != nil {
		return &UploadResponse{Success: false, Error: "Database error"}, nil
	}
	if existingFile != nil {
		return &UploadResponse{
			Success: true,
			ID:      existingFile.ID,
			Message: "File already exists",
			URL:     fmt.Sprintf("http://localhost:8080/image/%d", existingFile.ID),
		}, nil
	}

	if err := app.Storage.Save(newFilename, bytes.NewReader(fileBytes)); err != nil {
		return &UploadResponse{Success: false, Error: "Failed to save file"}, nil
	}

	newFile := models.ImageMetadata{
		Filename: newFilename,
		Format:   fileExt[1:], // Remove the leading dot
		Size:     header.Size,
	}
	if err := app.DB.Create(&newFile).Error; err != nil {
		return &UploadResponse{Success: false, Error: "Failed to save file metadata"}, nil
	}

	return &UploadResponse{
		Success: true,
		ID:      newFile.ID,
		Message: fmt.Sprintf("File %s uploaded successfully", header.Filename),
		URL:     fmt.Sprintf("http://localhost:8080/image/%d", newFile.ID),
	}, nil
}

func validateImage(header *multipart.FileHeader, maxUploadSize int64) error {
	if header.Size > maxUploadSize {
		return fmt.Errorf("the uploaded image is too big: %v. Please upload an image up to %v", header.Size, maxUploadSize)
	}

	contentType := strings.Split(header.Header.Get("Content-Type"), "/")[1]
	if !isSupportedFormat(contentType) {
		return fmt.Errorf("unsupported image format: %s, upload one of the following: %v", contentType, supportedFormats)
	}

	return nil
}

func isSupportedFormat(format string) bool {
	for _, f := range supportedFormats {
		if string(f) == format {
			return true
		}
	}
	return false
}

func generateFileHash(fileBytes []byte) string {
	hash := sha256.Sum256(fileBytes)
	return hex.EncodeToString(hash[:])
}

func checkExistingFile(db *gorm.DB, filename string) (*models.ImageMetadata, error) {
	var existingFile models.ImageMetadata
	result := db.Where("filename = ?", filename).First(&existingFile)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &existingFile, nil
}
