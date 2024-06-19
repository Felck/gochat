package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"syscall"

	"github.com/gorilla/websocket"
	"golang.org/x/sys/unix"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	conn    *websocket.Conn
	sendBuf chan []byte
}

var clients = make(map[*Client]bool)
var broadcast = make(chan []byte)

func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer conn.Close()

	client := &Client{conn: conn, sendBuf: make(chan []byte)}
	clients[client] = true

	go sendMessages(client)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			delete(clients, client)
			break
		}
		broadcast <- msg
	}
}

func sendMessages(client *Client) {
	for {
		msg := <-client.sendBuf
		if err := client.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("Error writing message: %v", err)
			client.conn.Close()
			delete(clients, client)
			break
		}
	}
}

func handleBroadcast() {
	for {
		msg := <-broadcast
		for client := range clients {
			select {
			case client.sendBuf <- msg:
			default:
				close(client.sendBuf)
				delete(clients, client)
			}
		}
	}
}

func main() {
	http.HandleFunc("/", handleConnection)
	go handleBroadcast()

	fmt.Println("Chat server started on :8080")
	server := http.Server{Addr: ":8080", Handler: nil}
	lc := net.ListenConfig{
		Control: func(network, address string, conn syscall.RawConn) error {
			var opErr error
			err := conn.Control(func(fd uintptr) {
				opErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
			})
			if err != nil {
				return err
			}
			return opErr
		},
	}

	listener, err := lc.Listen(context.Background(), "tcp", server.Addr)
	if err != nil {
		log.Fatal(err)
	}

	err = server.Serve(listener)

	if err != nil {
		log.Fatal(err)
	}
}
