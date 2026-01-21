// backend/config/config.go
package config

import (
	"log"
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

const (
	// cost of hashing algos should be between (11-14)
	BcryptCost = 12
)

var store *sessions.CookieStore

// here i init the auth from goth
// Note if you need other client id's you can use goth providers
func InitAuth() {
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	callbackURL := os.Getenv("BACKEND_URL") // this is the backend url

	// orginally i did this goth.providers((Google.Client, callbackURL))
	// key = "", maxAge:= 86400 *30, store:= sessions
	goth.UseProviders(
		google.New(googleClientID, googleClientSecret,
			callbackURL+"/api/auth/google/callback",
			"email", "profile"),
	)
	// my custom store
	gothic.Store = GetSessionStore()

	if callbackURL == "" {
		callbackURL = "http://localhost:8080"
	}

}

// get store comes from gorilla/sessions
func GetStore() *sessions.CookieStore {
	if store == nil {
		// look at the secure key session from the env file
		// to generate a sessionkey
		key := os.Getenv("SESSIONKEY")
		if key == "" {
			log.Println("SESSIONKEY not set")
			// log.Println("look at the sessionenv")
		}
		// the NewCookieStore() passing a secret key used to authenticate the session
		store = sessions.NewCookieStore([]byte(key))
		// As configured, this default store (gothic.Store) will generate cookies with Options
		store.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7, // will it be active for 7 days
			HttpOnly: true,
			Secure:   true,
		}
	}
	return store
}

// getter method that points to the sessions cookestore from max
func GetSessionStore() *sessions.CookieStore {
	return GetStore()
}

// Getter method for the frontend
func GetFrontendURL() string {
	url := os.Getenv("FRONTEND_URL")
	if url == "" {
		return "http://localhost:3000"
	}
	return url
}
