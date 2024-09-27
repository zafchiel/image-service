package main

import (
	"fmt"
	"net/http"

	"github.com/zafchiel/image-service/internal/config"
	"github.com/zafchiel/image-service/internal/handlers"
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

	server := http.Server{
		Addr:    cfg.ServerAddress,
		Handler: handlers.CreateRouter(app),
	}

	fmt.Println("Server is running on", cfg.ServerAddress)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
