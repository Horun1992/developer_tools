package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

const filePath = "version_history.json"

func ensureVersionHistoryFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		emptyJSON := []byte("[]")
		return os.WriteFile(path, emptyJSON, 0644)
	}
	return nil
}

func readVersionHistory() ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var versions []string
	err = json.Unmarshal(data, &versions)
	return versions, err
}

func writeVersionHistory(versions []string) error {
	updated, err := json.Marshal(versions)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, updated, 0644)
}

func handleGetVersionHistory(w http.ResponseWriter) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		w.Write([]byte("[]"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func handlePostVersionHistory(w http.ResponseWriter, r *http.Request) {
	var req struct{ Version string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Version == "" {
		http.Error(w, "empty version", http.StatusBadRequest)
		return
	}

	versions, err := readVersionHistory()
	if err != nil {
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
	if err := writeVersionHistory(versions); err != nil {
		http.Error(w, "failed to update version history", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleDeleteVersionHistory(w http.ResponseWriter, r *http.Request) {
	var req struct{ Version string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Version == "" {
		if err := os.WriteFile(filePath, []byte("[]"), 0644); err != nil {
			http.Error(w, "failed to clear version history", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	versions, err := readVersionHistory()
	if err != nil {
		http.Error(w, "cannot read version file", http.StatusInternalServerError)
		return
	}

	var updatedVersions []string
	for _, v := range versions {
		if v != req.Version {
			updatedVersions = append(updatedVersions, v)
		}
	}

	if err := writeVersionHistory(updatedVersions); err != nil {
		http.Error(w, "failed to update version history", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func VersionHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if err := ensureVersionHistoryFile(filePath); err != nil {
		log.Printf("❌ Не удалось создать файл: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handleGetVersionHistory(w)
	case http.MethodPost:
		handlePostVersionHistory(w, r)
	case http.MethodDelete:
		handleDeleteVersionHistory(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
