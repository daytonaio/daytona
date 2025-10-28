package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for testing
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Println("Client connected")

	// Wait a bit before closing
	time.Sleep(2 * time.Second)

	// Send close frame with code 4008
	closeMessage := websocket.FormatCloseMessage(4008, "Custom close code 4008")
	err = conn.WriteMessage(websocket.CloseMessage, closeMessage)
	if err != nil {
		log.Println("Write close error:", err)
		return
	}

	log.Println("Sent close code 4008 to client")

	// Wait a bit for graceful close
	time.Sleep(1 * time.Second)
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)

	port := ":8088"
	log.Printf("WebSocket server starting on %s", port)
	log.Println("Connect to: ws://localhost:8088/ws")

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}

