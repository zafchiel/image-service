package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/zafchiel/image-service/internal/models"
	"github.com/zafchiel/image-service/internal/session"
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
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mimeType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mimeType != "application/json" {
			http.Error(w, "Content-Type header must be application/json", http.StatusUnsupportedMediaType)
			return
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

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

	session, _ := session.Store.Get(r, session.Key)

	session.Values["user_id"] = user.ID

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
