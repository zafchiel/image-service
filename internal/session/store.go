package session

import (
	"github.com/gorilla/sessions"
)

var Store sessions.Store

const Key = "AUTH_SESSION_KEY"

func InitStore(secret string) {
	Store = sessions.NewCookieStore([]byte(secret))
}
