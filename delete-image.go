package main

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm/clause"
)

func deleteImageHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	var imageMetadata ImageMetadata
	result := db.Clauses(clause.Returning{}).Delete(&imageMetadata, id)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected == 0 {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	err := storage.Delete(imageMetadata.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"success": "true", "message": "Image deleted", "id": id})
}
