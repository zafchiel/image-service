package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/zafchiel/image-service/internal/config"
	"github.com/zafchiel/image-service/internal/handlers"
	"github.com/zafchiel/image-service/internal/middleware"
	"github.com/zafchiel/image-service/internal/models"
	"github.com/zafchiel/image-service/internal/storage"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	app := &handlers.App{
		DB:      db,
		Storage: storage.NewLocalStorage(cfg.StoragePath),
		Config:  cfg,
	}

	db.AutoMigrate(&models.ImageMetadata{}, &models.User{})

	router := http.NewServeMux()
	router.HandleFunc("POST /upload", handlers.NewUploadHandler(app).Handle)
	router.HandleFunc("GET /image/{id}", handlers.NewGetImageHandler(app).Handle)
	router.HandleFunc("DELETE /image/{id}", handlers.NewDeleteImageHandler(app).Handle)
	router.HandleFunc("POST /register", handlers.NewRegisterHandler(app).Handle)

	router.Handle("GET /docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))

	mdStack := middleware.Stack(
		middleware.Logger,
		middleware.NewRateLimiter(10, time.Second*10).Limit,
	)

	server := http.Server{
		Addr:    cfg.ServerAddress,
		Handler: mdStack(router),
	}

	fmt.Println("Server is running on", cfg.ServerAddress)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
