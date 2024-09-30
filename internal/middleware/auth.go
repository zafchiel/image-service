package middleware

import (
	"net/http"

	"github.com/zafchiel/image-service/internal/session"
)

const ASKey = "AUTH-SESSION-KEY"

// Check if user is authenticated
func AuthGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := session.Store.Get(r, ASKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if userID, ok := sess.Values["user_id"]; !ok || userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
