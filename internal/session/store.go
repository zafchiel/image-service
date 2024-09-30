package session

import (
	"github.com/gorilla/sessions"
)

var Store sessions.Store

const Key = "AUTH_SESSION_KEY"

func InitStore(secret string) {
	store := sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		HttpOnly: true,
	}

	Store = store
}
