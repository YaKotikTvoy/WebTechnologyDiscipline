package ws

import (
	"log"
	"sync"
	"time"
	"webchat/internal/models"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID uint
	Conn   *websocket.Conn
	Send   chan models.WSMessage
}

type MessageTracker struct {
	processed map[uint]time.Time
	mu        sync.RWMutex
}

func NewMessageTracker() *MessageTracker {
	return &MessageTracker{
		processed: make(map[uint]time.Time),
	}
}

func (mt *MessageTracker) IsProcessed(messageID uint) bool {
	mt.mu.RLock()
	processedTime, exists := mt.processed[messageID]
	mt.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Since(processedTime) > 5*time.Second {
		mt.mu.Lock()
		delete(mt.processed, messageID)
		mt.mu.Unlock()
		return false
	}

	return true
}

func (mt *MessageTracker) MarkProcessed(messageID uint) {
	mt.mu.Lock()
	mt.processed[messageID] = time.Now()
	mt.mu.Unlock()

	go func() {
		time.Sleep(10 * time.Second)
		mt.mu.Lock()
		if time.Since(mt.processed[messageID]) > 5*time.Second {
			delete(mt.processed, messageID)
		}
		mt.mu.Unlock()
	}()
}

type Hub struct {
	Clients      map[uint]*Client
	Register     chan *Client
	Unregister   chan *Client
	mu           sync.RWMutex
	messageTrack *MessageTracker
}

func NewHub() *Hub {
	return &Hub{
		Clients:      make(map[uint]*Client),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		messageTrack: NewMessageTracker(),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.UserID] = client
			h.mu.Unlock()
			go h.ReadPump(client)
			go h.WritePump(client)

		case client := <-h.Unregister:
			h.mu.Lock()
			if c, ok := h.Clients[client.UserID]; ok {
				close(c.Send)
				delete(h.Clients, client.UserID)
			}
			h.mu.Unlock()
		}
	}
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

		h.HandleMessage(client, msg)
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

func (h *Hub) HandleMessage(client *Client, msg models.WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	switch msg.Type {
	case "new_message":
		h.handleNewMessage(msg)
	case "message_read":
		h.handleMessageRead(msg)
	case "chat_invite":
		h.handleChatInvite(msg)
	case "chat_join_request":
		h.handleChatJoinRequest(msg)
	}
}

func (h *Hub) handleNewMessage(msg models.WSMessage) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		return
	}

	// Получаем ID сообщения
	var messageID uint = 0

	if msgData, ok := data["message"].(map[string]interface{}); ok {
		if id, ok := msgData["id"].(float64); ok {
			messageID = uint(id)
		}
	}

	if messageID == 0 {
		if id, ok := data["message_id"].(float64); ok {
			messageID = uint(id)
		}
	}

	if messageID > 0 {
		if h.messageTrack.IsProcessed(messageID) {
			log.Printf("Пропускаем дублированное сообщение ID: %d", messageID)
			return
		}
		h.messageTrack.MarkProcessed(messageID)
	}

	chatIDFloat, _ := data["chat_id"].(float64)
	senderIDFloat, _ := data["sender_id"].(float64)
	chatID := uint(chatIDFloat)
	senderID := uint(senderIDFloat)

	log.Printf("WebSocket: отправка сообщения %d в чат %d", messageID, chatID)

	chat, err := h.getChatFromRepository(chatID)
	if err != nil {
		return
	}

	for _, member := range chat.Members {
		if member.ID != senderID {
			h.SendToUser(member.ID, models.WSMessage{
				Type: "new_message",
				Data: data,
			})
		}
	}
}

func (h *Hub) handleMessageRead(msg models.WSMessage) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		return
	}

	readerID := uint(data["reader_id"].(float64))

	h.SendToUser(readerID, models.WSMessage{
		Type: "message_read_confirmation",
		Data: map[string]interface{}{
			"read_at": time.Now().Unix(),
		},
	})
}

func (h *Hub) handleChatInvite(msg models.WSMessage) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		return
	}

	userID := uint(data["user_id"].(float64))

	h.SendToUser(userID, models.WSMessage{
		Type: "chat_invite",
		Data: data,
	})
}

func (h *Hub) handleChatJoinRequest(msg models.WSMessage) {
}

func (h *Hub) getChatFromRepository(chatID uint) (*models.Chat, error) {
	var chat models.Chat
	return &chat, nil
}
