package main

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	connPool = make(map[*websocket.Conn]string)
	mutex    sync.Mutex
)

func BroadCastMessage(conn *websocket.Conn, msgToSend string, msgType int) (err error) {
	msgToSend = strings.ReplaceAll(msgToSend, "\r\n", "")
	for client := range connPool {
		if client != conn {
			// msgToSend := connPool[conn] + ": " + string(msg)
			if err := client.WriteMessage(msgType, []byte(msgToSend)); err != nil {
				log.Println("error broadcasting message")
				client.Close()
				mutex.Lock()
				delete(connPool, client)
				mutex.Unlock()
			}

		}
	}
	return
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error creating a connection object: %v", err)
	}
	defer conn.Close()
	mutex.Lock()
	connPool[conn] = ""
	mutex.Unlock()
	log.Println("New client connected")
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("error reading message from the client")
			msgToSend := connPool[conn] + ": Left the chat"
			BroadCastMessage(conn, msgToSend, websocket.TextMessage)
			mutex.Lock()
			delete(connPool, conn)
			mutex.Unlock()
			break
		}

		if connPool[conn] != "" {
			msgToSend := connPool[conn] + ": " + string(msg)
			BroadCastMessage(conn, msgToSend, msgType)
		} else {
			mutex.Lock()
			connPool[conn] = string(msg)
			mutex.Unlock()
			msgToSend := connPool[conn] + ": has joined the chat"
			BroadCastMessage(conn, msgToSend, websocket.TextMessage)
		}

	}

}
func main() {
	http.HandleFunc("/ws", HandleConnections)
	log.Println("About to start server")
	serverAddress := ":8080"
	log.Println("Websocket server started on: ", serverAddress)
	if err := http.ListenAndServe(serverAddress, nil); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
