package server

import (
	"developers_tools/internal/handlers"
	"developers_tools/internal/middleware"
	"net/http"
)

func RegisterRoutes() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Push routes
	http.HandleFunc("/push", serveFile("web/push/index.html"))
	http.HandleFunc("/push/script.js", serveFile("web/push/script.js"))
	http.HandleFunc("/send_push", middleware.BasicAuth(handlers.PushHandler))
	http.HandleFunc("/history.json", middleware.BasicAuth(handlers.HistoryHandler))
	http.HandleFunc("/logs/clear", middleware.BasicAuth(handlers.ClearLogHandler))
	http.Handle("/logs/", http.StripPrefix("/logs/", http.FileServer(http.Dir("logs"))))

	// Plate routes
	http.HandleFunc("/plate", serveFile("web/plate/index.html"))
	http.HandleFunc("/plate/script.js", serveFile("web/plate/script.js"))
	http.HandleFunc("/save_plate", middleware.BasicAuth(handlers.SavePlateHandler))
	http.HandleFunc("/plate_history", middleware.BasicAuth(handlers.PlateHistoryHandler))
	http.HandleFunc("/delete_plate", middleware.BasicAuth(handlers.DeletePlateHandler))
	http.HandleFunc("/version_history", middleware.BasicAuth(handlers.VersionHistoryHandler))
}

func serveFile(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}
