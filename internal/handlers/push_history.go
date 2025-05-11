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

func PushHistoryHandler(w http.ResponseWriter, r *http.Request) {
	historyLock.Lock()
	defer historyLock.Unlock()

	switch r.Method {
	case http.MethodGet:
		handlePushHistoryGet(w)
	case http.MethodDelete:
		handlePushHistoryDelete(w)
	case http.MethodPost:
		handlePushHistoryPost(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handlePushHistoryGet(w http.ResponseWriter) {
	data, err := os.ReadFile(historyFile)
	if err != nil {
		w.Write([]byte("[]"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func handlePushHistoryDelete(w http.ResponseWriter) {
	os.WriteFile(historyFile, []byte("[]"), 0644)
	w.WriteHeader(http.StatusOK)
}

func handlePushHistoryPost(w http.ResponseWriter, r *http.Request) {
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
