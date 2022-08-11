## websockety-mcchatface

Pub/sub in-memory websocket service that runs locally. It includes a server that accepts incoming websocket connections from multiple clients. It also includes a convenience client implementation to connect to the server and request a subscription to a topic. The server broadcasts incoming messages for the topic to all connected, subscribed clients.

### Running locally

#### 1. Start the server

Install the latest version of the server and start it (`-port` is optional, defaults to 8081):

```bash
go get github.com/karmakettle/websockety-mcchatface/cmd/server
go run cmd/server/main.go -port 8081
```

Output:
```
$ go run cmd/server/main.go -port 8081
2022/08/10 22:50:19 Starting server on port 8081
2022/08/10 22:50:19 Publish to clients subscribed to a topic via /publish. Example:
2022/08/10 22:50:19 curl -v -X POST -H 'Content-Type:application/json' http://localhost:8081/publish?topic=my_pipeline -d '{"test":"phase_1"}'
```

#### 2. Connect the clients

In a another terminal, install the latest version of the client and connect to the running server above. Specify a topic via `-topic` (required) and an optional port via `-port` (defaults to 8081):

```bash
go get github.com/karmakettle/websockety-mcchatface/cmd/client
go run cmd/client/main.go -topic my_pipeline -port 8081
```

Output:
```
$ go run cmd/client/main.go -topic my_pipeline -port 8081
2022/08/10 22:50:37 Connecting to ws://localhost:8081/subscribe?topic=my_pipeline
2022/08/10 22:50:37 {"subscription_status":"OK","topic":"my_pipeline"}
```

Subscribe more clients to the same topic if desired in separate terminals.

#### 3. Publish to clients

Send a POST request to `/publish` with the topic specified in the query parameters and the JSON data included in the body. Curl example:

```bash
curl -v -X POST -H 'Content-Type:application/json' \
  http://localhost:8081/publish?topic=my_pipeline -d '{"test":"phase_1"}'
```

Sample output from a connected client's perspective:
```
e$ go run cmd/client/main.go -topic my_pipeline -port 8081
2022/08/10 22:51:01 Connecting to ws://localhost:8081/subscribe?topic=my_pipeline
2022/08/10 22:51:01 {"subscription_status":"OK","topic":"my_pipeline"}
2022/08/10 22:51:09 {"test":"phase_1"}
```

### Run tests

```bash
go test -v ./...
```
