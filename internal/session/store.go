package session

import (
	"github.com/gorilla/sessions"
)

var Store sessions.Store

func InitStore(secret string) {
	Store = sessions.NewCookieStore([]byte(secret))
}
