//go:build !wasm

package tinysse

import (
	"strconv"
	"sync"
)

// SSEHub manages SSE clients and broadcasting.
type SSEHub struct {
	mu            sync.RWMutex
	clients       map[string]*clientConnection
	messageBuffer []SSEMessage
	config        *Config
	lastID        uint64
}

// NewHub creates a new SSEHub.
func NewHub(c *Config) *SSEHub {
	return &SSEHub{
		clients: make(map[string]*clientConnection),
		config:  c,
	}
}

// Broadcast sends a message to the specified channels.
func (h *SSEHub) Broadcast(data []byte, broadcast []string, handlerID uint8) {
	h.mu.Lock()
	h.lastID++
	msg := SSEMessage{
		ID:        strconv.FormatUint(h.lastID, 10),
		Data:      data,
		Targets:   broadcast,
		HandlerID: handlerID,
	}
	h.messageBuffer = append(h.messageBuffer, msg)

	// Trim buffer if it's too large
	if h.config.HistoryReplayBuffer > 0 && len(h.messageBuffer) > h.config.HistoryReplayBuffer {
		h.messageBuffer = h.messageBuffer[len(h.messageBuffer)-h.config.HistoryReplayBuffer:]
	}
	h.mu.Unlock()

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		for _, target := range broadcast {
			for _, channel := range client.Channels {
				if target == channel {
					client.Send <- msg
					break
				}
			}
		}
	}
}

// register adds a client to the hub.
func (h *SSEHub) register(client *clientConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.ID] = client
}

// unregister removes a client from the hub.
func (h *SSEHub) unregister(client *clientConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)
		close(client.Send)
	}
}

// GetMessagesSince returns all messages since the given ID.
func (h *SSEHub) GetMessagesSince(lastEventID string) []SSEMessage {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if lastEventID == "" {
		return nil
	}
	lastID, err := strconv.ParseUint(lastEventID, 10, 64)
	if err != nil {
		return nil
	}
	var messages []SSEMessage
	for _, msg := range h.messageBuffer {
		msgID, err := strconv.ParseUint(msg.ID, 10, 64)
		if err != nil {
			continue
		}
		if msgID > lastID {
			messages = append(messages, msg)
		}
	}
	return messages
}

// clientConnection represents a connected SSE client on the server side.
// Note: This is different from SSEClient which is the WASM client.
type clientConnection struct {
	ID       string
	UserID   string
	Role     string
	Channels []string
	Send     chan SSEMessage
}
