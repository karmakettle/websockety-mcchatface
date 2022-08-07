package main

import (
  "flag"
  "fmt"
  "net/http"
  "os"
)

func getTest(w http.ResponseWriter, r *http.Request) {
  // ctx := r.Context()

  topic := r.URL.Query().Get("topic")

  fmt.Printf("%s /test - topic %s\n", r.Method, topic)
}

func main() {
  var port string
  flag.StringVar(&port, "port", "8081", "Optionally provide the server port")
  flag.Parse()

  mux := http.NewServeMux()
  mux.HandleFunc("/test", getTest)

  fmt.Printf("Starting server on port %s\n", port)
  err := http.ListenAndServe(":" + port, mux)

  if err != nil {
    fmt.Printf("Server error: %s\n", err)
    os.Exit(1)
  }
}
