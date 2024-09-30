package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/zafchiel/image-service/internal/middleware"
	"github.com/zafchiel/image-service/internal/models"
)

type LoginHandler struct {
	app *App
}

func NewLoginHandler(app *App) *LoginHandler {
	return &LoginHandler{app: app}
}

type loginRequestBody struct {
	Email    string
	Password string
}

func (h *LoginHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var body loginRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Email == "" || body.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	um := models.NewUserModel(h.app.DB)
	user, err := um.LoginUser(body.Email, body.Password)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusBadRequest)
		return
	}

	session, _ := middleware.AuthStore.Get(r, middleware.ASKey)

	session.Values["user_id"] = user.ID

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
