package main

import (
  "flag"
  "fmt"
  "github.com/gorilla/websocket"
  "net/http"
  "os"
)

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

  fmt.Printf("%s /subscribe - topic %s\n", r.Method, topic)

  u := websocket.Upgrader{}
  c, err := u.Upgrade(w, r, nil)
  if err != nil {
    fmt.Println("Unable to establish connection: ", err)
    return
  }

  fmt.Println("websocket successfully created woo!", c.LocalAddr())
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

  // TODO require JSON body

  fmt.Printf("%s /publish - topic %s\n", r.Method, topic)

  // write to all the conns in the topic queues
  // if there's a failure, remove the cxn
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

