package handlers

import (
	"net/http"
	"time"

	"github.com/rs/cors"
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

	router.HandleFunc("POST /upload", NewUploadHandler(app).Handle)
	router.HandleFunc("GET /image/{id}", NewGetImageHandler(app).Handle)
	router.HandleFunc("DELETE /image/{id}", NewDeleteImageHandler(app).Handle)

	router.HandleFunc("POST /register", NewRegisterHandler(app).Handle)
	router.HandleFunc("POST /login", NewLoginHandler(app).Handle)

	router.Handle("GET /docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE"},
		AllowedHeaders: []string{"Content-Type"},
	})

	mdStack := middleware.Stack(
		middleware.Logger,
		middleware.NewRateLimiter(10, time.Second*10).Limit,
		corsHandler.Handler,
	)

	return mdStack(router)
}
