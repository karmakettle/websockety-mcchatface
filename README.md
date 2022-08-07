## websockety-mcchatface

Pub/sub in-memory websocket service, made up of a server and a client. The server accepts incoming connections from multiple clients. Incoming messages are broadcast to all clients subscribed to a topic in the request URL.

### Running locally

#### Server

Install the latest version:

```bash
go get github.com/karmakettle/websockety-mcchatface/cmd/server
```

Start the server with `go run` using an optional port flag (defaults to 8081)

```bash
go run cmd/server/main.go -port 8081
```

Example output:

```
TODO
```

#### Client

TODO

Endpoints:

```
TODO
```

### Run tests

TODO
