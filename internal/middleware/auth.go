package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

var AuthStore = sessions.NewCookieStore([]byte("secret-key"))

const ASKey = "AUTH-SESSION-KEY"

// Check if user is authenticated
func AuthGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := AuthStore.Get(r, ASKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if userID, ok := session.Values["user_id"]; !ok || userID == "" {
			fmt.Println(session.Values["user_id"])
			// http.Redirect(w, r, "/auth?type=login", 401)
			http.Error(w, "unauthorized", 401)
			return
		}

		next.ServeHTTP(w, r)
	})
}
