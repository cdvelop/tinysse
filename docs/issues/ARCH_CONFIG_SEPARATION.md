# Architecture: Configuration Separation

## Context
Currently, `tinysse` uses a single `Config` struct containing fields for both Server and Client. This causes:
1.  **Ambiguity**: Which fields are for whom?
2.  **Bloat**: WASM binaries include server configuration logic/fields (dead code).
3.  **Coupling**: Difficult to scale environments independently.

## Design Principles
- **SSE Transport Only**: tinysse handles ONLY the SSE protocol layer (connection, reconnection, message delivery).
- **No Authentication**: Security/auth is the responsibility of the consuming application.
- **No User Management**: Client identification/routing is external.
- **Raw Data Delivery**: The library delivers raw `[]byte` data without parsing.

## Proposed Design

### 1. `Config` (Shared)
Fields for **BOTH** environments. This struct is passed to `New()`.
```go
type Config struct {
    // Log is the centralized logger.
    // If nil, logging is disabled.
    Log func(args ...any)
}
```

### 2. `ServerConfig` (Build Tag: `!wasm`)
Fields strictly for the Server HTTP Handler.
```go
type ServerConfig struct {
    // ClientChannelBuffer prevents blocking on slow clients.
    // Recommended: 10-100.
    ClientChannelBuffer int
    
    // HistoryReplayBuffer manages the "Last-Event-ID" replay history.
    // Recommended: Depends on message frequency.
    HistoryReplayBuffer int
    
    // ChannelProvider resolves channels for each SSE connection.
    // If nil, a default provider is used that rejects all connections
    // with error "channel provider not configured".
    // See ARCH_CRUDP_INTEGRATION.md for details.
    ChannelProvider ChannelProvider
}
```

**Default ChannelProvider**: If `nil`, tinysse uses an internal default that returns error for all connections. This ensures the server works but clearly indicates misconfiguration. Tests can provide custom providers as needed.

### 3. `ClientConfig` (Build Tag: `wasm`)
Fields strictly for the Browser/WASM Client.
```go
type ClientConfig struct {
    // Endpoint is the SSE server URL.
    Endpoint string
    
    // RetryInterval in milliseconds for reconnection.
    RetryInterval int
    
    // MaxRetryDelay caps the exponential backoff.
    MaxRetryDelay int
    
    // MaxReconnectAttempts limits retry attempts. 0 = unlimited.
    MaxReconnectAttempts int
}
```

## Initialization Pattern

### Internal struct (private)
```go
// tinysse.go (shared, no build tag)
type tinySSE struct {
    config *Config
}
```

### Public API
```go
// tinysse.go (shared)
func New(c *Config) *tinySSE {
    return &tinySSE{config: c}
}

// server.go (!wasm)
func (t *tinySSE) Server(c *ServerConfig) *SSEServer {
    return &SSEServer{
        tinySSE: t,
        config:  c,
        hub:     newHub(c),
    }
}

// client.go (wasm)
func (t *tinySSE) Client(c *ClientConfig) *SSEClient {
    return &SSEClient{
        tinySSE: t,
        config:  c,
    }
}
```

### Usage Example
```go
// Server side (pkg/server.go)
server := tinysse.New(&tinysse.Config{
    Log: log.Println,
}).Server(&tinysse.ServerConfig{
    ClientChannelBuffer: 50,
    HistoryReplayBuffer: 100,
})

// Client side (web/client.go - WASM)
client := tinysse.New(&tinysse.Config{
    Log: console.Log,
}).Client(&tinysse.ClientConfig{
    Endpoint:     "/events",
    RetryInterval: 1000,
})
```

## Implementation Steps
1.  Rename `Config` fields, remove auth/user management.
2.  Create `config.go` (shared), `server_config.go` (!wasm), `client_config.go` (wasm).
3.  Make `tinySSE` struct private, `SSEServer` and `SSEClient` public.
4.  Both `SSEServer` and `SSEClient` embed `*tinySSE` for shared logging.
5.  Update build tags strictly.
