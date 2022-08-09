// Package client/main implements a client that subscribes to a topic on a locally running server.
// It receives messages from third parties that publish to the topic on the server.
// See server/main for the server implementation.
//
// Usage:
//
//	go run cmd/client/main.go [flags]
//
// Flags:
//
//	-topic
//	  Topic to receive incoming messages from (required).
//	-port
//	  Port the server is running on (optional, defaults to 8081).
package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)

	var topic string
	flag.StringVar(&topic, "topic", "", "Topic for subscription")

	var port string
	flag.StringVar(&port, "port", "8081", "Optionally provide the server port")
	flag.Parse()

	if topic == "" {
		log.Println("Topic required to set up subscription. Exiting")
		return
	}

	u := url.URL{Scheme: "ws", Host: "localhost:" + port, Path: "/subscribe", RawQuery: "topic=" + topic}
	log.Printf("Connecting to %s\n", u.String())

	// POST? https://github.com/gorilla/websocket/issues/689
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial:", err)
	}
	defer c.Close()

	// receive JSON messages
	for {
		var jsonMap interface{}
		err := c.ReadJSON(&jsonMap)
		if err != nil {
			log.Println("JSON error:", err)
		}

		dump(jsonMap)
	}
}

func dump(data interface{}) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshalling previously valid JSON, hmm...:", err)
	}
	log.Println(string(jsonBytes))
}
