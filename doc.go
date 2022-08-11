// Package server/main implements a server that handles incoming websocket
// connections and subscribes them to a given topic. The subscribed clients receive
// messages from third parties that publish to the topic. See client/main for the
// client implementation.
//
// Usage:
//
//	go run cmd/server/main.go [-port]
//
// The port flag is optional and defaults to 8081.
//
// Package client/main implements a client that subscribes to a topic on a locally
// running server. It receives messages from third parties that publish to the
// topic on the server. See server/main for the server implementation.
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
//
// package socketyutils // import "github.com/karmakettle/websockety-mcchatface/socketyutils"
//
// Package socketyutils provides validation, parsing, and logging utilties for the
// websockety clients, server, and tests.
//
// func Dump(data interface{}) string
// func GetValidTopic(w http.ResponseWriter, r *http.Request) (string, bool)
// func IsValidContentType(w http.ResponseWriter, r *http.Request, contentType string) bool
// func IsValidRequestMethod(w http.ResponseWriter, r *http.Request) bool
// func ParseJsonFromRequest(w http.ResponseWriter, r *http.Request) (map[string]interface{}, bool)
// func ReadJson(c *websocket.Conn) (interface{}, bool)
package doc
