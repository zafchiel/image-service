package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the original ResponseWriter
		wrappedWriter := &responseWriter{w, http.StatusOK}

		next.ServeHTTP(wrappedWriter, r)

		log.Println(wrappedWriter.statusCode, r.Method, r.URL.Path, time.Since(start))
	})
}
