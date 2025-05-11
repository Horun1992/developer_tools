package server

import (
	"developers_tools/internal/handlers"
	"developers_tools/internal/middleware"
	"net/http"
	"strings"
)

func MainHandler() {
	http.HandleFunc("/", hostRouter())
}

func hostRouter() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mux := http.NewServeMux()
		registerCommonRoutes(mux)

		switch {
		case strings.HasPrefix(r.Host, "push."):
			registerPushRoutes(mux)
		case strings.HasPrefix(r.Host, "plate."):
			registerPlateRoutes(mux)
		default:
			notFoundHandler(w, r)
			return
		}

		mux.ServeHTTP(w, r)
	}
}

func registerCommonRoutes(mux *http.ServeMux) {
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
}

func registerPushRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", serveFile("web/push/index.html"))
	mux.HandleFunc("/script.js", serveFile("web/push/script.js"))
	mux.HandleFunc("/send_push", middleware.BasicAuth(handlers.PushHandler))
	mux.HandleFunc("/history.json", middleware.BasicAuth(handlers.HistoryHandler))
	mux.HandleFunc("/logs/clear", middleware.BasicAuth(handlers.ClearLogHandler))
	mux.Handle("/logs/", http.StripPrefix("/logs/", http.FileServer(http.Dir("logs"))))
}

func registerPlateRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", serveFile("web/plate/index.html"))
	mux.HandleFunc("/script.js", serveFile("web/plate/script.js"))
	mux.HandleFunc("/save_plate", middleware.BasicAuth(handlers.SavePlateHandler))
	mux.HandleFunc("/plate_history", middleware.BasicAuth(handlers.PlateHistoryHandler))
	mux.HandleFunc("/delete_plate", middleware.BasicAuth(handlers.DeletePlateHandler))
	mux.HandleFunc("/version_history", middleware.BasicAuth(handlers.VersionHistoryHandler))
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func serveFile(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}
