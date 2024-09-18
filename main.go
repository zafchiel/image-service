package main

import (
	"fmt"
	"net/http"

	"github.com/zafchiel/image-service/middleware"
)

func main() {
	router := http.NewServeMux()
	router.HandleFunc("/", hello)

	server := http.Server{
		Addr:    ":8080",
		Handler: middleware.Logger(router),
	}

	fmt.Println("Server is running on port 8080")
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!")
}
