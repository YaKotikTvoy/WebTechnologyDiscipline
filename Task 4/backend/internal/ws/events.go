package ws

import (
	"sync"
	"time"
)

type EventManager struct {
	processedMessages sync.Map
	mu                sync.RWMutex
}

func NewEventManager() *EventManager {
	return &EventManager{}
}

func (em *EventManager) TrackMessage(messageID uint) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	key := messageID
	if _, exists := em.processedMessages.Load(key); exists {
		return false
	}

	em.processedMessages.Store(key, time.Now())

	go func() {
		time.Sleep(5 * time.Minute)
		em.processedMessages.Delete(key)
	}()

	return true
}
