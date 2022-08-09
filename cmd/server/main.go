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

// avoid read / write conflict from publish / subscribe endpoints
var topicsAndClients sync.Map

/*
  Requires POST, topic
*/
func subscribe(w http.ResponseWriter, r *http.Request) {
  //   if r.Method != http.MethodPost {
  //   err := r.Method + " not allowed for " + "/subscribe"
  //   fmt.Println(err)
  //   http.Error(w, err,  http.StatusMethodNotAllowed)
  //   return
  // }

  topic := r.URL.Query().Get("topic")
  if !(len(topic) > 0) {
    err := "No topic specified, unable to subscribe"
    fmt.Println(err)
    http.Error(w, err,  http.StatusBadRequest)
    return
  }

  //  maybe choose a reasonable buffer size
  //  ReadBufferSize:  1024,
  //  WriteBufferSize: 1024,
  u := websocket.Upgrader{}
  c, err := u.Upgrade(w, r, nil)
  if err != nil {
    fmt.Println("Unable to establish connection: ", err)
    return
  }

  clients, ok := topicsAndClients.Load(topic)
  var clientsSlice []*websocket.Conn
  if ok {
    // sync.Map returns type 'any', convert to slice to enable append
    clientsSlice = clients.([]*websocket.Conn)
  }

  clientsSlice = append(clientsSlice, c)
  topicsAndClients.Store(topic, clientsSlice)

  fmt.Printf("Currently serving %d connections\n", len(clientsSlice))

  // TODO: send as response
  fmt.Printf("Successfully subscribed to topic %s\n", topic)
}

/*
  Requires POST, topic, JSON body

  No authentication, no need to be subscribed to a topic
  before publishing. any entity can post any message to any
  existing topic

  Clients are only subscribed to one topic at a time
*/
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
    fmt.Printf("Failed to convert %s to JSON\n", body)
    return
  }

  clients, ok := topicsAndClients.Load(topic)
  if !ok {
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

