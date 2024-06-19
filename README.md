# gochat
This a minimal websocket broadcasting server written in Go.

## Server
The server automatically listens on `localhost:8080` and broadcasts incoming messages to all connected clients. 
```
cd server
go run main.go
```

## Client
The client automatically connects to `localhost:8080`, displays received messages and listens for console input. Type a message to send it to the server for broadcasting.
```
cd client
go run main.go
```
