## websockety-mcchatface

Pub/sub in-memory websocket service. Includes a server that accepts incoming websocket connections from multiple clients as well as a client that connects to the server to subscribe to a topic. The server broadcasts incoming messages for the topic to all subscribed clients.

### Running locally

#### 1. Start the server

Get the latest version of the server and start it (`-host` and `-port` are optional, they default to `localhost` and `8081`):

```bash
git clone https://github.com/karmakettle/websockety-mcchatface.git
cd websockety-mcchatface
# make sure your GOPATH is in a good state (or modify the .source_me file and run "source .source_me")
go run cmd/server/main.go -host localhost -port 8081
```

Output:
```
$ go run cmd/server/main.go -host localhost -port 8081
2022/08/11 00:56:24 Starting websocket server at localhost on port 8081
2022/08/11 00:56:24 Available at: ws://localhost:8081/subscribe?topic=my_topic
2022/08/11 00:56:24 Send to all connected clients on a given topic with: http://localhost:8081/publish?topic=my_favorite_topic
2022/08/11 00:56:24 Example:
2022/08/11 00:56:24 curl -v -X POST -H 'Content-Type:application/json' http://localhost:8081/publish?topic=my_pipeline -d '{"test":"phase_1"}'
```

#### 2. Connect the clients

In a another terminal, get the latest version of the client and connect to the running server above. Specify a topic via `-topic` (required) and an optional url and port via `-host` and `-port`. These default to `localhost` and `8081`.

```bash
go run cmd/client/main.go -topic my_pipeline -host localhost -port 8081
```

Output:
```
$ go run cmd/client/main.go -topic my_pipeline -host localhost -port 8081
2022/08/11 00:17:18 Connecting to ws://localhost:8081/subscribe?topic=my_pipeline
2022/08/11 00:17:18 {"subscription_status":"OK","topic":"my_pipeline"}
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
$ go run cmd/client/main.go -topic my_pipeline -host localhost -port 8081
2022/08/11 00:17:18 Connecting to ws://localhost:8081/subscribe?topic=my_pipeline
2022/08/11 00:17:18 {"subscription_status":"OK","topic":"my_pipeline"}
2022/08/11 00:19:45 {"test":"phase_1"}
```

### Run tests

```bash
go test -v ./...
```
