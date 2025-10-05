package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"xstudious-guide/authentication"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
}

type Hub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
}

var hub = Hub{
	clients:    make(map[string]*Client),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	broadcast:  make(chan []byte),
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("Client connected: %s", client.ID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				close(client.Send)
				delete(h.clients, client.ID)
				log.Printf("Client disconnected: %s", client.ID)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.ID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

func (h *Hub) SendTo(clientID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if client, ok := h.clients[clientID]; ok {
		client.Send <- message
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func serveWs(c *gin.Context) {
	tokenString := c.Query("token") // pass JWT in query string for simplicity
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	claims := authentication.ParseAccessToken(tokenString)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		ID:   claims.ID,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()
}

type EventMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// frontend javascript example:

// const ws = new WebSocket("ws://localhost:8080/ws?token=YOUR_JWT");

// ws.onopen = () => {
//   ws.send(JSON.stringify({ event: "ping" }));
//   ws.send(JSON.stringify({ event: "chat_message", data: "Hello backend!" }));
// };

// ws.onmessage = (msg) => {
//   console.log("Received:", msg.data);
// };

func (c *Client) readPump() {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
	}()
	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		var msg EventMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Println("invalid event format:", err)
			continue
		}

		log.Printf("recv from %s: %+v", c.ID, msg)

		switch msg.Event {
		case "chat_message":
			// Example: broadcast chat message
			if text, ok := msg.Data.(string); ok {
				hub.Broadcast([]byte(fmt.Sprintf("%s says: %s", c.ID, text)))
			}

		case "ping":
			// Respond only to this client
			c.Send <- []byte(`{"event":"pong"}`)

		case "join_room":
			// Handle joining rooms (you’d need to add room support in your Hub)
			log.Printf("%s joined room: %+v", c.ID, msg.Data)

		default:
			log.Printf("unknown event: %s", msg.Event)
		}
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
