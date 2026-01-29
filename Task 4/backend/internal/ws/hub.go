package ws

import (
	"log"
	"sync"
	"webchat/internal/models"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID uint
	Conn   *websocket.Conn
	Send   chan models.WSMessage
}

type Hub struct {
	Clients    map[uint]*Client
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uint]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
}

func (h *Hub) SendToUser(userID uint, message models.WSMessage) {
	h.mu.RLock()
	client, ok := h.Clients[userID]
	h.mu.RUnlock()

	if ok {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			h.mu.Lock()
			delete(h.Clients, userID)
			h.mu.Unlock()
		}
	}
}

func (h *Hub) SendToUsers(userIDs []uint, message models.WSMessage) {
	for _, userID := range userIDs {
		h.SendToUser(userID, message)
	}
}

func (h *Hub) ReadPump(client *Client) {
	defer func() {
		h.Unregister <- client
		client.Conn.Close()
	}()

	for {
		var msg models.WSMessage
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}
	}
}

func (h *Hub) WritePump(client *Client) {
	defer func() {
		client.Conn.Close()
	}()

	for {
		message, ok := <-client.Send
		if !ok {
			client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := client.Conn.WriteJSON(message)
		if err != nil {
			return
		}
	}
}
