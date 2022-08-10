// Package socketyutils provides validation, parsing, and logging utilties for the
// websockety clients, server, and tests.
package socketyutils

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
)

func ReadJson(c *websocket.Conn) (interface{}, bool) {
	var jsonMap interface{}
	err := c.ReadJSON(&jsonMap)
	if err != nil {
		log.Println("JSON error:", err)
		return nil, false
	}
	return jsonMap, true
}

func Dump(data interface{}) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshalling previously valid JSON, hmm...:", err)
	}
	return string(jsonBytes)
}

func IsValidRequestMethod(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		err := r.Method + " not allowed for " + "/publish"
		log.Println(err)
		http.Error(w, err, http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func GetValidTopic(w http.ResponseWriter, r *http.Request) (string, bool) {
	topic := r.URL.Query().Get("topic")
	if !(len(topic) > 0) {
		err := "No topic specified"
		log.Println(err)
		http.Error(w, err, http.StatusBadRequest)
		return "", false
	}
	return topic, true
}

func IsValidContentType(w http.ResponseWriter, r *http.Request, contentType string) bool {
	if contentType != "application/json" {
		err := "Invalid content type"
		log.Println(err)
		http.Error(w, err, http.StatusBadRequest)
		return false
	}
	return true
}

func ParseJsonFromRequest(w http.ResponseWriter, r *http.Request) (map[string]interface{}, bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read request body")
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return nil, false
	} else if string(body) == "" {
		log.Println("Empty body, unable to publish")
		http.Error(w, "Empty body, unable to publish", http.StatusBadRequest)
		return nil, false
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal([]byte(body), &jsonMap)
	if err != nil {
		log.Printf("Failed to convert %q to JSON\n", body)
		http.Error(w, "Invalid JSON: "+string(body), http.StatusBadRequest)
		return nil, false
	}

	return jsonMap, true
}
