package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	_ "os"
	"time"
)

var db *sql.DB

func InitPlateHandler(database *sql.DB) {
	db = database
}

func SavePlateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var plate struct {
		PlateID      string            `json:"plate_id"`
		Title        map[string]string `json:"title"`
		Body         map[string]string `json:"body"`
		Action       string            `json:"action"`
		ActionTitle  map[string]string `json:"action_title"`
		IconURL      string            `json:"icon_url"`
		WithCloseBtn bool              `json:"with_close_btn"`
		Conditions   map[string]string `json:"conditions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&plate); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	for _, field := range []struct {
		m map[string]string
	}{
		{plate.Title},
		{plate.Body},
		{plate.ActionTitle},
	} {
		if ru, ok := field.m["ru"]; ok {
			if _, exists := field.m["en"]; !exists {
				field.m["en"] = ru
			}
			if _, exists := field.m["tj"]; !exists {
				field.m["tj"] = ru
			}
		}
	}

	query := `
	INSERT INTO plates 
	(plate_id, title, body, action, action_title, icon_url, with_close_btn, conditions)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (plate_id) DO UPDATE SET
		title = EXCLUDED.title,
		body = EXCLUDED.body,
		action = EXCLUDED.action,
		action_title = EXCLUDED.action_title,
		icon_url = EXCLUDED.icon_url,
		with_close_btn = EXCLUDED.with_close_btn,
		conditions = EXCLUDED.conditions
	`

	titleJSON, _ := json.Marshal(plate.Title)
	bodyJSON, _ := json.Marshal(plate.Body)
	actionTitleJSON, _ := json.Marshal(plate.ActionTitle)
	conditionsJSON, _ := json.Marshal(plate.Conditions)

	_, err := db.Exec(query,
		plate.PlateID,
		titleJSON,
		bodyJSON,
		plate.Action,
		actionTitleJSON,
		plate.IconURL,
		plate.WithCloseBtn,
		conditionsJSON,
	)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{"status":"ok"}`))
}

func PlateHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query(`SELECT id, plate_id, title, body, action, action_title, icon_url, with_close_btn, conditions, created_at FROM plates ORDER BY created_at DESC LIMIT 50`)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []map[string]interface{}

	for rows.Next() {
		var id int
		var plateID, action, iconURL string
		var title, body, actionTitle, conditions []byte
		var withCloseBtn bool
		var createdAt time.Time

		err := rows.Scan(&id, &plateID, &title, &body, &action, &actionTitle, &iconURL, &withCloseBtn, &conditions, &createdAt)
		if err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"id":             id,
			"plate_id":       plateID,
			"action":         action,
			"icon_url":       iconURL,
			"with_close_btn": withCloseBtn,
			"conditions":     json.RawMessage(conditions),
			"created_at":     createdAt.Unix(),
			"title":          json.RawMessage(title),
			"body":           json.RawMessage(body),
			"action_title":   json.RawMessage(actionTitle),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func DeletePlateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PlateID string `json:"plate_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlateID == "" {
		http.Error(w, "invalid plate_id", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM plates WHERE plate_id = $1", req.PlateID)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
