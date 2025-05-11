package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var (
	historyFile = "history.json"
	historyLock sync.Mutex
)

// ClearLogHandler очищает лог-файл push.log
func ClearLogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	err := os.WriteFile(filepath.Join("logs", "push.log"), []byte{}, 0644)
	if err != nil {
		http.Error(w, "failed to clear log: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HistoryHandler обрабатывает историю Plate отправок
func HistoryHandler(w http.ResponseWriter, r *http.Request) {
	historyLock.Lock()
	defer historyLock.Unlock()

	switch r.Method {
	case http.MethodGet:
		data, err := os.ReadFile(historyFile)
		if err != nil {
			w.Write([]byte("[]"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)

	case http.MethodDelete:
		os.WriteFile(historyFile, []byte("[]"), 0644)
		w.WriteHeader(http.StatusOK)

	case http.MethodPost:
		var req struct {
			Delete int64 `json:"delete"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Delete == 0 {
			http.Error(w, "invalid delete timestamp", http.StatusBadRequest)
			return
		}
		var history []map[string]interface{}
		data, _ := os.ReadFile(historyFile)
		json.Unmarshal(data, &history)

		filtered := make([]map[string]interface{}, 0, len(history))
		for _, item := range history {
			if ts, ok := item["timestamp"].(float64); !ok || int64(ts) != req.Delete {
				filtered = append(filtered, item)
			}
		}
		updated, _ := json.Marshal(filtered)
		os.WriteFile(historyFile, updated, 0644)
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
