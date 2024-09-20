package main

import (
	"fmt"
	"net/http"

	"github.com/zafchiel/image-service/middleware"
)

const (
	port          = ":8080"
	maxUploadSize = 10 << 20 // 10 MB
)

func main() {
	router := http.NewServeMux()
	router.HandleFunc("GET /", hello)
	router.HandleFunc("POST /upload", uploadImage)
	router.HandleFunc("GET /image/{id}", getImage)

	server := http.Server{
		Addr:    port,
		Handler: middleware.Logger(router),
	}

	fmt.Println("Server is running on port", port)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!")
}
