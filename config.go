package tinysse

// Config holds the configuration for TinySSE.
type Config struct {
	// ClientChannelBuffer defines the size of the Go channel for each connected client.
	// This prevents the server from blocking if a single client is slow to receive messages.
	// A full buffer will cause the server to drop messages for that specific client or block depending on implementation.
	// Recommended: 10-100.
	ClientChannelBuffer int

	// HistoryReplayBuffer defines the number of recent messages to keep in memory.
	// These messages are used to replay missed events when a client reconnects using the Last-Event-ID header.
	// Recommended: Depends on message frequency and required reliability strictly for reconnection.
	HistoryReplayBuffer int

	Endpoint             string
	RetryInterval        int
	MaxRetryDelay        int
	MaxReconnectAttempts int
	AllowedOrigins       []string

	// Callbacks
	OnConnect    func(clientID string)
	OnDisconnect func(clientID string)
	OnMessage    func(msg *SSEMessage)
	OnError      func(err error)

	// Auth
	TokenValidator func(token string) (userID, role string, err error)
	TokenProvider  func() (token string, err error)
}
