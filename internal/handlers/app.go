package handlers

import (
	"net/http"
	"time"

	"github.com/zafchiel/image-service/internal/config"
	"github.com/zafchiel/image-service/internal/middleware"
	"github.com/zafchiel/image-service/internal/storage"
	"gorm.io/gorm"
)

type App struct {
	DB      *gorm.DB
	Config  *config.Config
	Storage storage.Storage
}

func CreateRouter(app *App) http.Handler {
	router := http.NewServeMux()
	router.Handle("POST /upload",
		middleware.AuthGuard(
			http.HandlerFunc(NewUploadHandler(app).Handle),
		),
	)
	router.Handle("GET /image/{id}",
		middleware.AuthGuard(
			http.HandlerFunc(NewGetImageHandler(app).Handle),
		),
	)
	router.Handle("DELETE /image/{id}",
		middleware.AuthGuard(
			http.HandlerFunc(NewDeleteImageHandler(app).Handle),
		),
	)
	router.HandleFunc("POST /register", NewRegisterHandler(app).Handle)
	router.HandleFunc("POST /login", NewLoginHandler(app).Handle)

	router.Handle("GET /docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))

	mdStack := middleware.Stack(
		middleware.Logger,
		middleware.NewRateLimiter(10, time.Second*10).Limit,
	)

	return mdStack(router)
}
