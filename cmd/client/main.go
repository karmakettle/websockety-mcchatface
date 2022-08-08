package main

import (
  "flag"
  "fmt"
  "github.com/gorilla/websocket"
  "log"
  "net/url"
  "os"
)

/*
  A client that subscribes to a given topic (required) by
  establishing a persistent connection to a locally running
  server listening on the given port (defaults to 8081).

  While subscribed (while the connection is open), the client
  receives all messages published by a third party to the given
  topic.
*/
func main() {
  var topic string
  flag.StringVar(&topic, "topic", "", "Topic for subscription")

  var port string
  flag.StringVar(&port, "port", "8081", "Optionally provide the server port")
  flag.Parse()

  if topic == "" {
     fmt.Println("Topic required to set up subscription. Exiting")
     return
  }

  u := url.URL{Scheme: "ws", Host: "localhost:" + port, Path: "/subscribe", RawQuery: "topic=" + topic}
  fmt.Printf("Connecting to %s\n", u.String())

  // POST?
  // https://github.com/gorilla/websocket/issues/689
  c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
  if err != nil {
    fmt.Println("Dial:", err)
    os.Exit(1)
  }
  defer c.Close()

  // receive messages
  for {
    _, message, err := c.ReadMessage()
    if err != nil {
      fmt.Println("Read error:", err)
      return
    }
    log.Printf("%s", message)
  }
}
