package middleware

import (
	"log"
	"net/http"
)

var username = ""
var password = ""

func InitMiddleware(user string, pass string) {
	username = user
	password = pass
}

func BasicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		log.Printf("🔒 BasicAuth: env user='%s' pass='%s'\n", username, pass)
		log.Printf("🔒 BasicAuth: selected user='%s' pass='%s'\n", user, pass)
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
