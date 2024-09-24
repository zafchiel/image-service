package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/zafchiel/image-service/middleware"
	bolt "go.etcd.io/bbolt"
)

type ImageFormat string

const (
	JPG  ImageFormat = "jpg"
	JPEG ImageFormat = "jpeg"
	PNG  ImageFormat = "png"
)

var supportedFormats = []ImageFormat{JPG, JPEG, PNG}
var db *bolt.DB

func main() {
	var err error
	db, err = bolt.Open("image.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	// Create the bucket if it doesn't exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("images"))
		if err != nil {
			return fmt.Errorf("error creating bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error creating bucket:", err)
		return
	}

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
