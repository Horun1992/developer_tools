package handlers

import (
	"bytes"
	handlers "developers_tools/internal/utils"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2/google"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	projectIDs = map[string]string{
		"release": "kurbisomoni",
		"debug":   "converter-somoni-debug",
	}
	debugKeyFile   = "firebase-debug.json"
	releaseKeyFile = "firebase-release.json"
)

func PushHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	buildType, _ := req["build_type"].(string)
	if buildType == "" {
		buildType = "debug"
	}
	projectID := projectIDs[buildType]
	var keyFile string
	if buildType == "release" {
		keyFile = releaseKeyFile
	} else {
		keyFile = debugKeyFile
	}

	creds, err := os.ReadFile(keyFile)
	if err != nil {
		http.Error(w, "failed to read credentials: "+err.Error(), 500)
		return
	}
	conf, err := google.JWTConfigFromJSON(creds, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		http.Error(w, "failed to parse credentials: "+err.Error(), 500)
		return
	}
	token, err := conf.TokenSource(r.Context()).Token()
	if err != nil {
		http.Error(w, "failed to get token: "+err.Error(), 500)
		return
	}

	data := make(map[string]string)
	if rawData, ok := req["data"].(map[string]interface{}); ok {
		for k, v := range rawData {
			if k == "title" || k == "body" || k == "sound" {
				continue
			}
			switch val := v.(type) {
			case string:
				data[k] = val
			default:
				encoded, _ := json.Marshal(val)
				data[k] = string(encoded)
			}
		}
	}

	var titleMap, bodyMap map[string]string
	if raw, ok := req["title"].(string); ok && raw != "" {
		_ = json.Unmarshal([]byte(raw), &titleMap)
	}
	if raw, ok := req["body"].(string); ok && raw != "" {
		_ = json.Unmarshal([]byte(raw), &bodyMap)
	}

	if ru, ok := titleMap["ru"]; ok {
		if _, exists := titleMap["en"]; !exists {
			titleMap["en"] = ru
		}
		if _, exists := titleMap["tj"]; !exists {
			titleMap["tj"] = ru
		}
	}
	if ru, ok := bodyMap["ru"]; ok {
		if _, exists := bodyMap["en"]; !exists {
			bodyMap["en"] = ru
		}
		if _, exists := bodyMap["tj"]; !exists {
			bodyMap["tj"] = ru
		}
	}

	successTopics := []string{}
	var lastResp map[string]interface{}

	var conditions map[string]string
	if condRaw, ok := data["conditions"]; ok && condRaw != "" && condRaw != "{}" {
		_ = json.Unmarshal([]byte(condRaw), &conditions)
	}

	var topics []string
	if versionsStr, ok := conditions["version"]; ok && versionsStr != "" {
		for _, version := range strings.Split(versionsStr, ",") {
			version = strings.TrimSpace(version)
			for _, lang := range []string{"ru", "en", "tj"} {
				topics = append(topics, fmt.Sprintf("%s_%s", version, lang))
			}
		}
	} else {
		topics = []string{"ru", "en", "tj"}
	}

	for _, topic := range topics {
		parts := strings.Split(topic, "_")
		lang := topic
		if len(parts) == 2 {
			lang = parts[1]
		}
		localTitle := titleMap[lang]
		localBody := bodyMap[lang]

		notif := map[string]interface{}{
			"channel_id": "GENERAL",
		}
		if s, ok := req["data"].(map[string]interface{})["sound"]; ok {
			str, _ := s.(string)
			if str != "none" && str != "" {
				data["sound"] = str
				notif["sound"] = str
			}
		}

		priority := "high"
		if p, ok := data["priority"]; ok {
			priority = p
		}

		message := map[string]interface{}{
			"message": map[string]interface{}{
				"topic": topic,
				"notification": map[string]interface{}{
					"title": localTitle,
					"body":  localBody,
					"image": data["image"],
				},
				"android": map[string]interface{}{
					"priority":     priority,
					"notification": notif,
				},
				"data": data,
			},
		}

		msgBody, _ := json.Marshal(message)
		fmt.Printf("📦 Отправка в топик %s:\n%s\n", topic, msgBody)

		url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", projectID)
		httpReq, _ := http.NewRequest("POST", url, io.NopCloser(bytes.NewReader(msgBody)))
		httpReq.Header.Set("Authorization", "Bearer "+token.AccessToken)
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(httpReq)
		if resp != nil {
			defer resp.Body.Close()
		}

		if err == nil && resp.StatusCode == 200 {
			successTopics = append(successTopics, topic)
		}
		json.NewDecoder(resp.Body).Decode(&lastResp)

		msgID, _ := lastResp["name"].(string)
		entry := map[string]interface{}{
			"timestamp":  time.Now().Unix(),
			"payload":    req,
			"success":    err == nil && resp.StatusCode == 200,
			"topic":      topic,
			"title":      localTitle,
			"message_id": msgID,
		}
		handlers.SaveHistoryEntry(entry)
	}

	w.Header().Set("Content-Type", "application/json")
	if len(successTopics) > 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"topics":  successTopics,
			"message": lastResp["name"],
		})
	} else {
		w.WriteHeader(500)
		errMsg := "Unknown error"
		status := "FAILED"
		if errMap, ok := lastResp["error"].(map[string]interface{}); ok {
			status, _ = errMap["status"].(string)
			errMsg, _ = errMap["message"].(string)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"status": status,
			"error":  errMsg,
		})
	}
}
