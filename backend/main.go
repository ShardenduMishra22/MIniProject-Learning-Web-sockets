package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	// github.com/gorilla/mux: Used for routing.

	"github.com/gorilla/websocket"
	// github.com/gorilla/websocket: Provides WebSocket handling capabilities.
)


// websocket.Upgrader configures how WebSocket connections are handled. 
// Here, it allows connections from all origins (CheckOrigin: true).
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// When a client connects, handleConnections handles the upgrade to WebSocket 
// and adds the client to a map of connected clients.
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

var clients = make(map[*Client]bool)
var broadcast = make(chan []byte)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}
	client := &Client{conn: conn, send: make(chan []byte)}
	clients[client] = true

	defer func() {
		delete(clients, client)
		client.conn.Close()
	}()

	go clientWriter(client)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		broadcast <- msg
	}
}

func clientWriter(client *Client) {
	for msg := range client.send {
		err := client.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			fmt.Println("Error writing message:", err)
			break
		}
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			client.send <- msg
		}
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ws", handleConnections)

	go handleMessages()

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", r)
}