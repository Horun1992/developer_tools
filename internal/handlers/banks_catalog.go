package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

var db *sql.DB

func InitPlateHandler(database *sql.DB) {
	db = database
}

// BanksHandler returns the list of banks from banks_catalog table
func BanksHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM banks_catalog")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var results []map[string]interface{}
	for rows.Next() {
		// Prepare a slice for column values and pointers
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}

		// Scan the row into pointers
		if err := rows.Scan(pointers...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Build a map for the row
		rowMap := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			val := values[i]
			// Convert []byte to string for JSON readability
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		results = append(results, rowMap)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
