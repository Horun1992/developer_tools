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

		if !ok || user == "" || pass == "" || user == "null" || pass == "null" || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Printf("✅ BasicAuth: authorized user='%s'\n", user)
		handler(w, r)
	}
}
