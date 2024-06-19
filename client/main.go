package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

var input = make(chan string)
var die = make(chan struct{})

func recieve(conn *websocket.Conn) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			close(die)
			break
		}
		fmt.Printf("Message: %s/n", msg)
	}
}

func readCLI() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input <- scanner.Text()
	}
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080", nil)
	if err != nil {
		log.Fatal("Error dialing:", err)
	}
	defer conn.Close()

	go recieve(conn)
	go readCLI()

	for {
		select {
		case msg := <-input:
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				log.Printf("Error writing message: %v", err)
				return
			}
		case <-interrupt:
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error closing socket:", err)
				return
			}
		case <-die:
			return
		}
	}
}
