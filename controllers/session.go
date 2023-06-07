package controllers

import (
	"github.com/gorilla/sessions"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	Key   = []byte("8IJ$&MN98LKM129KJ0P£**FCCCFG180") // Key for session cookie
	Store = sessions.NewCookieStore(Key)              // Create session cookie store
)

func init() {
	Store.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 0,
	}
}
