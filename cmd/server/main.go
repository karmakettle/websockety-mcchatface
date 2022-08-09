// Package server/main implements a server that handles incoming websocket connections and subscribes them to a given topic.
// The subscribed clients receive messages from third parties that publish to the topic.
// See client/main for the client implementation.
//
// Usage:
//
//	go run cmd/server/main.go [-port]
//
// The port flag is optional and defaults to 8081.
package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

// Thread-safe map of topics and an array of subscribed clients.
var topicsAndClients sync.Map

func main() {
	log.SetOutput(os.Stdout)

	var port string
	flag.StringVar(&port, "port", "8081", "Optionally provide the server port")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/subscribe", subscribe)
	mux.HandleFunc("/publish", publish)

	log.Printf("Starting server on port %s\n", port)
	err := http.ListenAndServe(":"+port, mux)

	if err != nil {
		log.Fatal("Server error: %s\n", err)
	}
}

// Subscribe is an http handler that accepts incoming websocket connections
// and subscribes them to the topic specified in the `topic` query parameter.
// This is accomplished by adding the topic, client to the topicsAndClients map.
// Clients are only subscribed to one topic at a time.
func subscribe(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	if isValid := isValidTopic(w, r, topic); !isValid { return }

	u := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	c, err := u.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to establish connection: ", err)
		return
	}

	// get existing clients list for the topic or create a new one
	clients, topicFound := topicsAndClients.Load(topic)
	subscribeClient(c, topic, clients, topicFound)

	if err = c.WriteJSON(map[string]string{"subscription_status": "OK", "topic": topic}); err != nil {
		log.Println("Subscription confirmation failed, closing connection")
		c.Close()
	}
}

// Publish is an http handler that sends JSON data in the incoming request to all connected clients for the topic specified in the `topic` query parameter.
// The topic must exist in the topicsAndClients map.
// The publisher itself doesn't need to be subscribed to the topic.
func publish(w http.ResponseWriter, r *http.Request) {
	if isValid := isValidRequestMethod(w, r); !isValid { return }

	topic := r.URL.Query().Get("topic")
	if isValid := isValidTopic(w, r, topic); !isValid { return }

	contentType := r.Header.Get("Content-Type")
	if isValid := isValidContentType(w, r, contentType); !isValid { return }

	requestJson, ok := parseJsonFromRequest(w, r); if !ok { return }

	// verify topic exists, get subscribed clients
	clients, topicFound := topicsAndClients.Load(topic)
	if !topicFound {
		http.Error(w, "Topic \""+topic+"\" doesn't exist, unable to publish", http.StatusBadRequest)
		return
	}

	broadcastMessageAndUpdateClients(topic, requestJson, clients)
}

/////////////////////////////////////////////////////////////////////////////
// HELPERS
/////////////////////////////////////////////////////////////////////////////

// TODO - docs
func subscribeClient(c *websocket.Conn, topic string, clients any, topicFound bool) {
	var clientsSlice []*websocket.Conn
	if topicFound {
		// sync.Map returns type 'any', convert to slice to enable append
		clientsSlice = clients.([]*websocket.Conn)
	}

	clientsSlice = append(clientsSlice, c)
	topicsAndClients.Store(topic, clientsSlice)
}

// TODO - docs
func broadcastMessageAndUpdateClients(topic string, requestJson map[string]interface{}, clients any) {
	// sync.Map returns type 'any', convert to slice to enable indexing
	var clientsSlice []*websocket.Conn
	clientsSlice = clients.([]*websocket.Conn)

	// publish to all clients subscribed to the topic
	clientsCopy := clientsSlice[:0]
	for _, client := range clientsSlice {
		if err := client.WriteJSON(requestJson); err == nil {
			clientsCopy = append(clientsCopy, client)
		} else {
			// client disconnect detected after two failed write attempts
			log.Println(err)
		}
	}

	topicsAndClients.Store(topic, clientsCopy)
}

/////////////////////////////////////////////////////////////////////////////
// UTILS
/////////////////////////////////////////////////////////////////////////////

func isValidRequestMethod(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		err := r.Method + " not allowed for " + "/publish"
		log.Println(err)
		http.Error(w, err, http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func isValidTopic(w http.ResponseWriter, r *http.Request, topic string) bool {
	if !(len(topic) > 0) {
		err := "No topic specified"
		log.Println(err)
		http.Error(w, err, http.StatusBadRequest)
		return false
	}
	return true
}

func isValidContentType(w http.ResponseWriter, r *http.Request, contentType string) bool {
	if contentType != "application/json" {
		err := "Invalid content type"
		log.Println(err)
		http.Error(w, err, http.StatusBadRequest)
		return false
	}
	return true
}

func parseJsonFromRequest(w http.ResponseWriter, r *http.Request) (map[string]interface{}, bool) {
	body, err := ioutil.ReadAll(r.Body)
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
		log.Printf("Failed to convert \"%s\" to JSON\n", body)
		http.Error(w, "Invalid JSON: "+string(body), http.StatusBadRequest)
		return nil, false
	}

	return jsonMap, true
}
