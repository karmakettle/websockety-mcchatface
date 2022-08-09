package socketyutils

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
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
