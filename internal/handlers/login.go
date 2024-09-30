package handlers

import (
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

func (h *LoginHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	um := models.NewUserModel(h.app.DB)
	user, err := um.LoginUser(email, password)
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
