package middleware

import (
	"log"
	"net/http"
	"os"
)

var username = os.Getenv("DEV_USER")
var password = os.Getenv("DEV_PASS")

func BasicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			log.Println("🔒 BasicAuth: no auth header provided")
		} else if user == "" || pass == "" {
			log.Println("🔒 BasicAuth: empty username or password")
		} else if user == "null" || pass == "null" {
			log.Println("🔒 BasicAuth: user or pass == 'null'")
		} else if user != username || pass != password {
			log.Printf("🔒 BasicAuth: invalid credentials user='%s'\n", user)
		}

		if !ok || user == "" || pass == "" || user == "null" || pass == "null" || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Printf("✅ BasicAuth: authorized user='%s'\n", user)
		handler(w, r)
	}
}
