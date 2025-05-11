package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func ensureVersionHistoryFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		emptyJSON := []byte("[]")
		return os.WriteFile(path, emptyJSON, 0644)
	}
	return nil
}

func VersionHistoryHandler(w http.ResponseWriter, r *http.Request) {
	const filePath = "version_history.json"

	if err := ensureVersionHistoryFile(filePath); err != nil {
		log.Printf("❌ Не удалось создать файл: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		data, err := os.ReadFile(filePath)
		if err != nil {
			w.Write([]byte("[]"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	case http.MethodPost:
		var req struct{ Version string }
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if req.Version == "" {
			http.Error(w, "empty version", 400)
			return
		}
		var versions []string
		data, _ := os.ReadFile(filePath)
		if err := json.Unmarshal(data, &versions); err != nil {
			http.Error(w, "invalid version history data", http.StatusInternalServerError)
			return
		}

		for _, v := range versions {
			if v == req.Version {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		versions = append(versions, req.Version)
		updated, _ := json.Marshal(versions)
		os.WriteFile(filePath, updated, 0644)
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
