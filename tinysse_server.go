//go:build !wasm

package tinysse

// New initializes a new TinySSE instance for the server.
// New initializes a new TinySSE instance for the server.
func New(c *Config) *TinySSE {
	return &TinySSE{
		config: c,
		hub:    NewHub(c),
	}
}

// TinySSE is the main struct for the library (Server-side).
type TinySSE struct {
	config *Config
	hub    *SSEHub
}

// Broadcast sends a message to the specified channels.
func (s *TinySSE) Broadcast(data []byte, broadcast []string, handlerID uint8) {
	s.hub.Broadcast(data, broadcast, handlerID)
}
