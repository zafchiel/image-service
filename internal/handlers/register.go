package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zafchiel/image-service/internal/errors"
	"github.com/zafchiel/image-service/internal/models"
)

type RegisterHandler struct {
	app *App
}

func NewRegisterHandler(app *App) *RegisterHandler {
	return &RegisterHandler{app: app}
}

func (h *RegisterHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")

	if email == "" || username == "" || password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	um := models.UserModel{DB: h.app.DB}
	newUser, err := um.InsertUser(email, username, password)
	if err != nil {
		if err == errors.ErrEmailInUse {
			http.Error(w, "Email already in use", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"success": "true",
		"message": "User registered successfully",
		"id":      strconv.Itoa(int(newUser.ID)),
	})
}
