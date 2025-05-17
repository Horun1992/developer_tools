package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	keyTitle      = "title"
	keyBody       = "body"
	keyImage      = "image"
	keyPriority   = "priority"
	keySound      = "sound"
	keyConditions = "conditions"
)

const (
	LangRU = "ru"
	LangEN = "en"
	LangTJ = "tj"
)

var (
	projectIDs = map[string]string{
		"release": "kurbisomoni",
		"debug":   "converter-somoni-debug",
	}
	debugKeyFile   = "firebase-debug.json"
	releaseKeyFile = "firebase-release.json"
)

func SendPushHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handlePushPost(w, r)
	default:
		handlePushNotAllowed(w)
	}
}

func handlePushPost(w http.ResponseWriter, r *http.Request) {
	req, projectID, token, data, titleMap, bodyMap, err := preparePushRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	topics := generateTopics(data)
	successTopics := []string{}
	var lastResp map[string]interface{}

	for _, topic := range topics {
		localTitle := titleMap[getLang(topic)]
		localBody := bodyMap[getLang(topic)]

		message := buildMessage(topic, localTitle, localBody, data)
		msgBody, _ := json.Marshal(message)

		fmt.Printf("📦 Отправка в топик %s:\n%s\n", topic, msgBody)
		resp, err := sendFCMRequest(projectID, token.AccessToken, msgBody)
		if resp != nil {
			defer resp.Body.Close()
		}

		if err == nil && resp.StatusCode == 200 {
			successTopics = append(successTopics, topic)
		}
		json.NewDecoder(resp.Body).Decode(&lastResp)

		msgID, _ := lastResp["name"].(string)
		SaveHistoryEntry(map[string]interface{}{
			"timestamp":  time.Now().Unix(),
			"payload":    req,
			"success":    err == nil && resp.StatusCode == 200,
			"topic":      topic,
			"title":      localTitle,
			"message_id": msgID,
		})
	}

	writePushResponse(w, successTopics, lastResp)
}

func preparePushRequest(r *http.Request) (map[string]interface{}, string, *oauth2.Token, map[string]string, map[string]string, map[string]string, error) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, "", nil, nil, nil, nil, err
	}

	buildType, _ := req["build_type"].(string)
	if buildType == "" {
		buildType = "debug"
	}
	projectID := projectIDs[buildType]
	keyFile := debugKeyFile
	if buildType == "release" {
		keyFile = releaseKeyFile
	}

	creds, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, "", nil, nil, nil, nil, fmt.Errorf("failed to read credentials: %v", err)
	}
	conf, err := google.JWTConfigFromJSON(creds, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return nil, "", nil, nil, nil, nil, fmt.Errorf("failed to parse credentials: %v", err)
	}
	token, err := conf.TokenSource(r.Context()).Token()
	if err != nil {
		return nil, "", nil, nil, nil, nil, fmt.Errorf("failed to get token: %v", err)
	}

	data := extractPushData(req)
	titleMap, _ := decodeLangMap(req[keyTitle])
	bodyMap, _ := decodeLangMap(req[keyBody])
	normalizeLanguages(titleMap)
	normalizeLanguages(bodyMap)

	return req, projectID, token, data, titleMap, bodyMap, nil
}

func extractPushData(req map[string]interface{}) map[string]string {
	data := make(map[string]string)
	if rawData, ok := req["data"].(map[string]interface{}); ok {
		for k, v := range rawData {
			switch val := v.(type) {
			case string:
				data[k] = val
			default:
				encoded, _ := json.Marshal(val)
				data[k] = string(encoded)
			}
		}
	}
	return data
}

func decodeLangMap(raw any) (map[string]string, error) {
	res := make(map[string]string)
	if str, ok := raw.(string); ok && str != "" {
		if err := json.Unmarshal([]byte(str), &res); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func normalizeLanguages(m map[string]string) {
	if ru, ok := m[LangRU]; ok {
		if _, exists := m[LangEN]; !exists {
			m[LangEN] = ru
		}
		if _, exists := m[LangTJ]; !exists {
			m[LangTJ] = ru
		}
	}
}

func generateTopics(data map[string]string) []string {
	var conditions map[string]string
	if condRaw, ok := data[keyConditions]; ok && condRaw != "" && condRaw != "{}" {
		_ = json.Unmarshal([]byte(condRaw), &conditions)
	}
	var topics []string
	if versionsStr, ok := conditions["version"]; ok && versionsStr != "" {
		for _, version := range strings.Split(versionsStr, ",") {
			version = strings.TrimSpace(version)
			for _, lang := range []string{LangRU, LangEN, LangTJ} {
				topics = append(topics, fmt.Sprintf("%s_%s", version, lang))
			}
		}
	} else {
		topics = []string{LangRU, LangEN, LangTJ}
	}
	return topics
}

func getLang(topic string) string {
	parts := strings.Split(topic, "_")
	if len(parts) == 2 {
		return parts[1]
	}
	return topic
}

func buildMessage(topic, title, body string, data map[string]string) map[string]interface{} {
	notif := map[string]interface{}{"channel_id": "GENERAL"}
	if s, ok := data[keySound]; ok {
		if s == "" || s == "none" {
			delete(data, keySound)
		} else {
			notif[keySound] = s
		}
	}
	priority := "high"
	if p, ok := data[keyPriority]; ok {
		priority = p
	}

	data[keyTitle] = title // Нужен для аналитки
	data[keyBody] = body   // Нужен для аналитки

	return map[string]interface{}{
		"message": map[string]interface{}{
			"topic": topic,
			"notification": map[string]interface{}{
				"title": title,
				"body":  body,
				"image": data[keyImage],
			},
			"android": map[string]interface{}{
				"priority":     priority,
				"notification": notif,
			},
			"data": data,
		},
	}
}

func sendFCMRequest(projectID, token string, body []byte) (*http.Response, error) {
	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", projectID)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func writePushResponse(w http.ResponseWriter, successTopics []string, lastResp map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if len(successTopics) > 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"topics":  successTopics,
			"message": lastResp["name"],
		})
	} else {
		w.WriteHeader(http.StatusInternalServerError)
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

func handlePushNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
