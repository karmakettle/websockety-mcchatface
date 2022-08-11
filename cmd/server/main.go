// Package server/main implements a server that handles incoming websocket connections
// and subscribes them to a given topic. The subscribed clients receive messages from
// third parties that publish to the topic. See client/main for the client implementation.
//
// Usage:
//
//	go run cmd/server/main.go [-port]
//
// The port flag is optional and defaults to 8081.
package main

import (
	"flag"
	"github.com/gorilla/websocket"
	sutils "github.com/karmakettle/websockety-mcchatface/socketyutils"
	"log"
	"net/http"
	"os"
	"sync"
)

// Thread-safe map of topics and an array of subscribed clients.
var topicsAndClients sync.Map

func main() {
	log.SetOutput(os.Stdout)

	var host string
	flag.StringVar(&host, "host", "localhost", "Optionally provide the server url")

	var port string
	flag.StringVar(&port, "port", "8081", "Optionally provide the server port")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/subscribe", subscribe)
	mux.HandleFunc("/publish", publish)

	log.Printf("Starting websocket server at %s on port %s\n", host, port)
	log.Printf("Available at: ws://%s:%s/subscribe?topic=my_topic", host, port)
	log.Printf("Send to all connected clients on a given topic with: http://%s:%s/publish?topic=my_favorite_topic", host, port)
	log.Println("Example:")
	log.Printf("curl -v -X POST -H 'Content-Type:application/json' http://%s:%s/publish?topic=my_pipeline -d '{\"test\":\"phase_1\"}'", host, port)
	err := http.ListenAndServe(host+":"+port, mux)

	if err != nil {
		log.Fatalf("Server error: %s\n", err)
	}
}

/////////////////////////////////////////////////////////////////////////////
// HANDLERS
/////////////////////////////////////////////////////////////////////////////

// Subscribe is an http handler that accepts incoming websocket connections
// and subscribes them to the topic specified in the `topic` query parameter.
// This is accomplished by adding the topic, client to the topicsAndClients map.
// Clients are only subscribed to one topic at a time.
func subscribe(w http.ResponseWriter, r *http.Request) {
	topic, isValid := sutils.GetValidTopic(w, r)
	if !isValid {
		return
	}

	u := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	c, err := u.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to establish connection: ", err)
		return
	}

	// race condition created if this write happens after subscribeClient()
	err1 := c.WriteJSON(map[string]string{"subscription_status": "OK", "topic": topic})
	if err1 != nil {
		log.Println(err1)
	}

	// get existing clients list for the topic or create a new one
	clients, topicFound := topicsAndClients.Load(topic)
	subscribeClient(c, topic, clients, topicFound)
}

// Publish is an http handler that sends JSON data in the incoming request to all connected
// clients for the topic specified in the `topic` query parameter. The topic must exist in
// the topicsAndClients map. The publisher itself doesn't need to be subscribed to the topic.
func publish(w http.ResponseWriter, r *http.Request) {
	isValid := sutils.IsValidRequestMethod(w, r)
	if !isValid {
		return
	}

	topic, isValid1 := sutils.GetValidTopic(w, r)
	if !isValid1 {
		return
	}

	contentType := r.Header.Get("Content-Type")
	isValid2 := sutils.IsValidContentType(w, r, contentType)
	if !isValid2 {
		return
	}

	requestJson, ok := sutils.ParseJsonFromRequest(w, r)
	if !ok {
		return
	}

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

// SubscribeClient converts clients subscribed to the given topic into an array
// and adds the incoming websocket.Conn to the list of subscribed clients for that
// topic. If the topic doesn't already exist, it's created. The topicsAndClients map
// is updated with the new topic and/or subscription.
func subscribeClient(c *websocket.Conn, topic string, clients any, topicFound bool) {
	var clientsSlice []*websocket.Conn
	if topicFound {
		// sync.Map returns type 'any', convert to slice to enable append
		clientsSlice = clients.([]*websocket.Conn)
	}

	clientsSlice = append(clientsSlice, c)
	topicsAndClients.Store(topic, clientsSlice)
}

// BroadcastMessageAndUpdateClients attempts to write the specified requestJson to the list of clients from topicsAndClients for the given topic.
// If a broken pipe is detected, the client is removed from the list, and the topicsAndClients map is updated.
func broadcastMessageAndUpdateClients(topic string, requestJson map[string]interface{}, clients any) {
	// sync.Map returns type 'any', convert to slice to enable iteration
	var clientsSlice []*websocket.Conn
	clientsSlice = clients.([]*websocket.Conn)

	// keep track of healthy clients
	healthyClients := make([]*websocket.Conn, 0)
	for _, client := range clientsSlice {
		err := client.WriteJSON(requestJson)
		if err != nil {
			// don't inlude broken client in updated list
			// this error (broken pipe) only happens on the second call to WriteJSON after the disconnect...
			// TODO - detect sooner and clean up before trying to publish to clients?
			log.Println(err)
			continue
		}
		healthyClients = append(healthyClients, client)
	}

	topicsAndClients.Store(topic, healthyClients)
}
