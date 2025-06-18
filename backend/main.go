package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Client struct {
    conn     *websocket.Conn
    username string
}

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[*Client]bool)
var broadcast = make(chan Message)

type Message struct {
    Type    string `json:"type"`    // "message", "join", "leave"
    Sender  string `json:"sender"`  // ユーザー名
    Content string `json:"content"` // メッセージ本文（Type == "message" のとき）
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("WebSocket upgrade error:", err)
        return
    }
    defer ws.Close()

    _, usernameBytes, err := ws.ReadMessage()
    if err != nil {
        log.Println("Failed to read username:", err)
        return
    }

    username := string(usernameBytes)
    client := &Client{conn: ws, username: username}
    clients[client] = true

    // 入室通知
    broadcast <- Message{Type: "join", Sender: username}

    for {
        _, msg, err := ws.ReadMessage()
        if err != nil {
            delete(clients, client)
            // 退出通知
            broadcast <- Message{Type: "leave", Sender: username}
            break
        }

        broadcast <- Message{
            Type:    "message",
            Sender:  client.username,
            Content: string(msg),
        }
    }
}

func handleMessages() {
    for {
        msg := <-broadcast
        payload, err := json.Marshal(msg)
        if err != nil {
            log.Println("JSON marshal error:", err)
            continue
        }

        for client := range clients {
            err := client.conn.WriteMessage(websocket.TextMessage, payload)
            if err != nil {
                log.Printf("Write error for %s: %v", client.username, err)
                client.conn.Close()
                delete(clients, client)
            }
        }
    }
}

func main() {
    fs := http.FileServer(http.Dir("./../static"))
    http.Handle("/", fs)
    http.HandleFunc("/ws", handleConnections)

    go handleMessages()

    fmt.Println("サーバー起動: http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
