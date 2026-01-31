package ws

import (
	"fmt"
	"log"
	"sync"
	"time"
	"webchat/internal/models"

	"github.com/gorilla/websocket"
)

var (
	globalMessageTracker = sync.Map{}
	trackerMutex         sync.RWMutex
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
	case "chat_deleted":
		h.handleChatDeleted(msg)
	case "chat_created":
		h.handleChatCreated(msg)
	}
}

func (h *Hub) handleNewMessage(msg models.WSMessage) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		return
	}

	messageID := uint(data["message_id"].(float64))
	chatID := uint(data["chat_id"].(float64))
	senderID := uint(data["sender_id"].(float64))

	key := fmt.Sprintf("msg_%d_%d_%d", chatID, messageID, time.Now().Unix()/60)
	if _, exists := globalMessageTracker.Load(key); exists {
		log.Printf("Пропускаем дублированное сообщение ID: %d в чате %d", messageID, chatID)
		return
	}
	globalMessageTracker.Store(key, true)

	go func() {
		time.Sleep(65 * time.Second)
		globalMessageTracker.Delete(key)
	}()

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

func (h *Hub) handleChatCreated(msg models.WSMessage) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		return
	}

	chatID := uint(data["chat_id"].(float64))
	userID := uint(data["user_id"].(float64))

	trackerKey := getMessageTrackerKey("chat_created", chatID, 0)
	if isMessageProcessed(trackerKey) {
		return
	}
	markMessageAsProcessed(trackerKey)

	h.SendToUser(userID, models.WSMessage{
		Type: "chat_created",
		Data: data,
	})
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

func (h *Hub) handleChatDeleted(msg models.WSMessage) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		return
	}

	chatID := uint(data["chat_id"].(float64))

	chat, err := h.getChatFromRepository(chatID)
	if err != nil {
		return
	}

	for _, member := range chat.Members {
		h.SendToUser(member.ID, models.WSMessage{
			Type: "chat_deleted",
			Data: data,
		})
	}
}

func (h *Hub) getChatFromRepository(chatID uint) (*models.Chat, error) {
	var chat models.Chat
	return &chat, nil
}

func getMessageTrackerKey(msgType string, chatID, messageID uint) string {
	return string(rune(messageID)) + "_" + msgType + "_" + string(rune(chatID))
}

func isMessageProcessed(key string) bool {
	trackerMutex.RLock()
	defer trackerMutex.RUnlock()

	if timestamp, exists := globalMessageTracker.Load(key); exists {
		if time.Since(timestamp.(time.Time)) < 5*time.Second {
			return true
		}
	}
	return false
}

func markMessageAsProcessed(key string) {
	trackerMutex.Lock()
	globalMessageTracker.Store(key, time.Now())
	trackerMutex.Unlock()

	go func() {
		time.Sleep(10 * time.Second)
		trackerMutex.Lock()
		globalMessageTracker.Delete(key)
		trackerMutex.Unlock()
	}()
}
