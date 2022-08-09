// Package server/main implements a server that handles incoming websocket connections and subscribes them to a given topic.
// The subscribed clients receive messages from third parties that publish to the topic.
// See client/main for the client implementation.
//
// Usage:
//
//   go run cmd/server/main.go [-port]
//
// The port flag is optional and defaults to 8081.
package main

import (
  "encoding/json"
  "flag"
  "fmt"
  "github.com/gorilla/websocket"
  "io/ioutil"
  "net/http"
  "os"
  "sync"
)

// Thread-safe map of topics and an array of subscribed clients.
var topicsAndClients sync.Map

// Subscribe is an http handler that accepts incoming websocket connections
// and subscribes them to the topic specified in the `topic` query parameter.
// This is accomplished by adding the topic, client to the topicsAndClients map.
// Clients are only subscribed to one topic at a time.
func subscribe(w http.ResponseWriter, r *http.Request) {
  topic := r.URL.Query().Get("topic")
  if !(len(topic) > 0) {
    err := "No topic specified, unable to subscribe"
    fmt.Println(err)
    http.Error(w, err,  http.StatusBadRequest)
    return
  }

  u := websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
  }
  c, err := u.Upgrade(w, r, nil)
  if err != nil {
    fmt.Println("Failed to establish connection: ", err)
    return
  }

  // get existing clients list for the topic or create a new one
  clients, topicFound := topicsAndClients.Load(topic)
  var clientsSlice []*websocket.Conn
  if topicFound {
    // sync.Map returns type 'any', convert to slice to enable append
    clientsSlice = clients.([]*websocket.Conn)
  }

  clientsSlice = append(clientsSlice, c)
  topicsAndClients.Store(topic, clientsSlice)

  // TODO: logging, debug level
  // fmt.Printf("Topic %s currently serving %d connections\n", topic, len(clientsSlice))

  // might be nice to have server-side logging for this too with some kind of session id
  c.WriteMessage(websocket.TextMessage, []byte("Successfully subscribed to topic \"" + topic + "\""))
}

// Publish is an http handler that sends JSON data in the incoming request to all connected clients for the topic specified in the `topic` query parameter.
// The topic must exist in the topicsAndClients map.
// The publisher itself doesn't need to be subscribed to the topic.
func publish(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodPost {
    err := r.Method + " not allowed for " + "/publish"
    fmt.Println(err)
    http.Error(w, err,  http.StatusMethodNotAllowed)
    return
  }

  topic := r.URL.Query().Get("topic")
  if !(len(topic) > 0) {
    err := "No topic specified, unable to publish"
    fmt.Println(err)
    http.Error(w, err, http.StatusBadRequest)
    return
  }

  // TODO: check content header to verify application/json?

  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    fmt.Println("Failed to read request body")
    return
  } else if string(body) == "" {
    fmt.Println("Empty body, unable to publish")
    return
  }

  var jsonMap map[string]interface{}
  err = json.Unmarshal([]byte(body), &jsonMap)
  if err != nil {
    fmt.Printf("Failed to convert \"%s\" to JSON\n", body)
    return
  }

  // verify topic exists, get subscribed clients
  clients, topicFound := topicsAndClients.Load(topic)
  if !topicFound {
    http.Error(w, "Topic \"" + topic + "\" doesn't exist, unable to publish", http.StatusBadRequest)
    return
  }

  // sync.Map returns type 'any', convert to slice to enable indexing
  var clientsSlice []*websocket.Conn
  clientsSlice = clients.([]*websocket.Conn)

  // publish to all clients subscribed to the topic
  clientsCopy := clientsSlice[:0]
  for _, client := range clientsSlice {
    if err = client.WriteJSON(jsonMap); err == nil {
      clientsCopy = append(clientsCopy, client)
    } else {
      // client disconnect detected after two failed write attempts
      fmt.Println(err)
    }
  }

  topicsAndClients.Store(topic, clientsCopy)
}

func main() {
  var port string
  flag.StringVar(&port, "port", "8081", "Optionally provide the server port")
  flag.Parse()

  mux := http.NewServeMux()
  mux.HandleFunc("/publish", publish)
  mux.HandleFunc("/subscribe", subscribe)

  fmt.Printf("Starting server on port %s\n", port)
  err := http.ListenAndServe(":" + port, mux)

  if err != nil {
    fmt.Printf("Server error: %s\n", err)
    os.Exit(1)
  }
}
