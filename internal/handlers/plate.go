package handlers

import (
	"encoding/json"
	"net/http"
	_ "os"
	"time"
)

func SavePlateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var plate struct {
		PlateID string            `json:"plate_id"`
		Title   map[string]string `json:"title"`
		Body    map[string]string `json:"body"`
		Action  struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"action"`
		ActionTitle          map[string]string `json:"action_title"`
		IconURL              string            `json:"icon_url"`
		WithCloseBtn         bool              `json:"with_close_btn"`
		WithReopenBtn        bool              `json:"with_reopen_btn"`
		Conditions           map[string]string `json:"conditions"`
		IsActive             bool              `json:"is_active"`
		ChangedAtMs          int64             `json:"changed_at_ms"`
		ShouldHideAfterClick bool              `json:"should_hide_after_click"`
	}

	if err := json.NewDecoder(r.Body).Decode(&plate); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	// If only is_active and changed_at_ms are provided, update those fields only
	if plate.Title == nil && plate.Body == nil && plate.ActionTitle == nil && plate.Action.Type == "" && plate.Action.Value == "" && plate.IconURL == "" && !plate.WithCloseBtn && !plate.WithReopenBtn && len(plate.Conditions) == 0 {
		plate.ChangedAtMs = time.Now().UTC().UnixNano() / int64(time.Millisecond)
		// Execute partial update
		_, err := db.Exec(
			"UPDATE plates SET is_active=$1, changed_at_ms=$2 WHERE plate_id=$3",
			plate.IsActive, plate.ChangedAtMs, plate.PlateID,
		)
		if err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(`{"status":"ok"}`))
		return
	}

	// Заполняем отсутствующие языки
	for _, field := range []map[string]string{plate.Title, plate.Body, plate.ActionTitle} {
		if ru, ok := field["ru"]; ok {
			if _, exists := field["en"]; !exists {
				field["en"] = ru
			}
			if _, exists := field["tj"]; !exists {
				field["tj"] = ru
			}
		}
	}

	createdAt := time.Now().UTC()
	changedAtMs := createdAt.UnixNano() / int64(time.Millisecond)

	query := `
    INSERT INTO plates (
        plate_id,
        title,
        body,
        action,
        action_title,
        icon_url,
        with_close_btn,
        with_reopen_btn,
        is_active,
        conditions,
        created_at,
        changed_at_ms,
    	should_hide_after_click
    ) VALUES (
        $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13
    )
    ON CONFLICT (plate_id) DO UPDATE SET
        title           = EXCLUDED.title,
        body            = EXCLUDED.body,
        action          = EXCLUDED.action,
        action_title    = EXCLUDED.action_title,
        icon_url        = EXCLUDED.icon_url,
        with_close_btn  = EXCLUDED.with_close_btn,
        with_reopen_btn = EXCLUDED.with_reopen_btn,
        is_active       = EXCLUDED.is_active,
        conditions      = EXCLUDED.conditions,
        changed_at_ms   = EXCLUDED.changed_at_ms,
        should_hide_after_click = EXCLUDED.should_hide_after_click
    `

	titleJSON, _ := json.Marshal(plate.Title)
	bodyJSON, _ := json.Marshal(plate.Body)
	actionTitleJSON, _ := json.Marshal(plate.ActionTitle)
	conditionsJSON, _ := json.Marshal(plate.Conditions)
	actionJSON, _ := json.Marshal(plate.Action)

	_, err := db.Exec(
		query,
		plate.PlateID,
		titleJSON,
		bodyJSON,
		actionJSON,
		actionTitleJSON,
		plate.IconURL,
		plate.WithCloseBtn,
		plate.WithReopenBtn,
		plate.IsActive,
		conditionsJSON,
		createdAt,
		changedAtMs,
		plate.ShouldHideAfterClick,
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

	rows, err := db.Query(`SELECT id, plate_id, title, body, action, action_title, icon_url, with_close_btn, with_reopen_btn, is_active, conditions, created_at, should_hide_after_click FROM plates ORDER BY created_at DESC`)
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
		var withCloseBtn, withReopenBtn, isActive, shouldHideAfterClick bool
		var createdAt time.Time

		err := rows.Scan(&id, &plateID, &title, &body, &action, &actionTitle, &iconURL, &withCloseBtn, &withReopenBtn, &isActive, &conditions, &createdAt, &shouldHideAfterClick)
		if err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"id":                      id,
			"plate_id":                plateID,
			"action":                  json.RawMessage(action),
			"icon_url":                iconURL,
			"with_close_btn":          withCloseBtn,
			"with_reopen_btn":         withReopenBtn,
			"is_active":               isActive,
			"conditions":              json.RawMessage(conditions),
			"created_at":              createdAt.Unix(),
			"title":                   json.RawMessage(title),
			"body":                    json.RawMessage(body),
			"action_title":            json.RawMessage(actionTitle),
			"should_hide_after_click": shouldHideAfterClick,
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
