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
		panic("failed to connect database: " + err.Error())
	}

	app := &handlers.App{
		DB:      db,
		Storage: storage.NewLocalStorage(cfg.StoragePath),
		Config:  cfg,
	}

	if err := db.AutoMigrate(&models.ImageMetadata{}, &models.User{}); err != nil {
		panic("failed to run auto migrations: " + err.Error())
	}

	server := http.Server{
		Addr:    cfg.ServerAddress,
		Handler: handlers.CreateRouter(app),
	}

	fmt.Println("Server is running on", cfg.ServerAddress)
	if err := server.ListenAndServe(); err != nil {
		panic("failed to start server: " + err.Error())
	}
}
