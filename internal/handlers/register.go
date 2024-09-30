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

type registerRequestBody struct {
	Username string
	Email    string
	Password string
}

func (h *RegisterHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var body registerRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if body.Email == "" || body.Username == "" || body.Password == "" {
		http.Error(w, "All fields are required: username, email, password", http.StatusBadRequest)
		return
	}

	um := models.UserModel{DB: h.app.DB}
	newUser, err := um.InsertUser(body.Email, body.Username, body.Password)
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
