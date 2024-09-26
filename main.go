package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/zafchiel/image-service/middleware"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ImageFormat string

const (
	JPG  ImageFormat = "jpg"
	JPEG ImageFormat = "jpeg"
	PNG  ImageFormat = "png"
)

var supportedFormats = []ImageFormat{JPG, JPEG, PNG}
var maxUploadSize int64 = 10 << 20 // 10 MB
var storage Storage = &LocalStorage{root: "assets"}

type ImageMetadata struct {
	gorm.Model
	Filename string
	Format   string
	Size     int64
}

var db *gorm.DB

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("image.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&ImageMetadata{})

	router := http.NewServeMux()
	router.HandleFunc("GET /", hello)
	router.HandleFunc("POST /upload", uploadImage)
	router.HandleFunc("GET /image/{id}", getImage)
	router.Handle("GET /docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))

	mdStack := middleware.Stack(
		middleware.Logger,
		// 10 requests per 10 seconds
		middleware.NewRateLimiter(10, time.Second*10).Limit,
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}

	server := http.Server{
		Addr:    port,
		Handler: mdStack(router),
	}

	fmt.Println("Server is running on port", port)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!")
}
