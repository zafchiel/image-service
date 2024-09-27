package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/zafchiel/image-service/internal/errors"
	"github.com/zafchiel/image-service/internal/models"
	"gorm.io/gorm/clause"
)

type DeleteImageHandler struct {
	app *App
}

func NewDeleteImageHandler(app *App) *DeleteImageHandler {
	return &DeleteImageHandler{app: app}
}

func (h *DeleteImageHandler) Handle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, errors.ErrInvalidID.Error(), http.StatusBadRequest)
		return
	}

	var imageMetadata models.ImageMetadata
	result := h.app.DB.Clauses(clause.Returning{}).Delete(&imageMetadata, id)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected == 0 {
		http.Error(w, errors.ErrImageNotFound.Error(), http.StatusNotFound)
		return
	}

	err := h.app.Storage.Delete(imageMetadata.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"success": "true", "message": "Image deleted", "id": id})
}
