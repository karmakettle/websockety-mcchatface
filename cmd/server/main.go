package main

import (
  "flag"
  "fmt"
  "net/http"
  "os"
)

/*
  Requires POST, topic
*/
func subscribe(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodPost {
    err := r.Method + " not allowed for " + "/subscribe"
    fmt.Println(err)
    http.Error(w, err,  http.StatusMethodNotAllowed)
    return
  }

  topic := r.URL.Query().Get("topic")
  if !(len(topic) > 0) {
    err := "No topic specified, unable to subscribe"
    fmt.Println(err)
    http.Error(w, err,  http.StatusBadRequest)
    return
  }

  fmt.Printf("%s /subscribe - topic %s\n", r.Method, topic)
}

/*
  Requires POST, topic, JSON body
*/
func publish(w http.ResponseWriter, r *http.Request) {
  // websocket connection should already be established via /subscribe

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

