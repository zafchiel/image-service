package middleware

import (
	"net/http"

	"github.com/zafchiel/image-service/internal/session"
)

// Check if user is authenticated
func AuthGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := session.Store.Get(r, session.Key)
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
