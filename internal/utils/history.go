package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
)

var (
	historyFile = "history.json"
	historyLock sync.Mutex
)

func ClearLogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	err := os.WriteFile("logs/push.log", []byte{}, 0644)
	if err != nil {
		http.Error(w, "failed to clear log: "+err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

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
			http.Error(w, "invalid delete timestamp", 400)
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

func SaveHistoryEntry(entry map[string]interface{}) {
	historyLock.Lock()
	defer historyLock.Unlock()

	var history []map[string]interface{}
	if data, err := os.ReadFile(historyFile); err == nil {
		json.Unmarshal(data, &history)
	}

	history = append(history, entry)
	updated, _ := json.Marshal(history)
	os.WriteFile(historyFile, updated, 0644)
}
