# Architecture: Cleanup & Verification

## Context
After implementing Config Separation and Client Decoupling, we need to ensure:
1. No dead code in WASM binary.
2. All tests pass.
3. API is clean and consistent.

## Cleanup Checklist

### 1. Remove Dead Code
- [ ] Delete `hub_wasm.go` (empty SSEHub placeholder).
- [ ] Remove `encoding/json` import from client.go.
- [ ] Remove unused fields from old `Config` struct.
- [ ] Delete old `config.go` after migration.
- [ ] Remove `SSEError` struct from models.go (use `tinystring.Err()` instead - Q23).

### 2. Rename for Clarity
- [x] Rename `SSEClient` in `hub.go` to `clientConnection` (server-side client representation). ✅ DONE
- [ ] Ensure `SSEClient` (public) is only the WASM client.
- [ ] Ensure `SSEServer` (public) is only the server handler.

### 3. File Structure After Refactor
```
tinysse/
├── tinysse.go          # Shared: tinySSE struct, New(), Config
├── models.go           # Shared: SSEMessage only (Q22 - used by both sides)
├── server.go           # !wasm: SSEServer, ServeHTTP
├── server_config.go    # !wasm: ServerConfig
├── hub.go              # !wasm: Hub, clientConnection
├── interfaces_server.go # !wasm: ChannelProvider, SSEPublisher (Q14)
├── client.go           # wasm: SSEClient, Connect, Close, OnMessage, OnError
├── client_config.go    # wasm: ClientConfig
└── (tests)
```

### 4. Build Tag Verification
- [ ] `tinysse.go` - NO build tag (shared).
- [ ] `models.go` - NO build tag (shared) - SSEMessage only, no SSEError.
- [ ] `server.go` - `//go:build !wasm`
- [ ] `server_config.go` - `//go:build !wasm`
- [ ] `hub.go` - `//go:build !wasm`
- [ ] `interfaces_server.go` - `//go:build !wasm`
- [ ] `client.go` - `//go:build wasm`
- [ ] `client_config.go` - `//go:build wasm`

### 5. WASM Binary Size Verification
Before refactor:
- [ ] Measure WASM size with `tinygo build -o test.wasm -target wasm ./...`

After refactor:
- [ ] Measure again and compare.
- [ ] Expected reduction: Removal of `encoding/json` should reduce ~50-100KB.

### 6. Dependency Audit
- [ ] `models.go` should NOT import `tinystring` (SSEMessage is a simple struct).
- [ ] `tinystring.Err()` is used in server/client code, not in models.go.
- [ ] WASM files should only import: `syscall/js`, internal packages.
- [ ] Server files can import: `net/http`, `sync`, `crypto/rand`, etc.

## Test Requirements (Post-Refactor)

**Testing Strategy**: Use `test.sh` for real tests (stdlib + WASM with wasmbrowsertest).

### Server Tests (`server_test.go`) - Run with stdlib
1. `TestNewServer` - Verify `New().Server()` returns valid `*SSEServer`.
2. `TestServerBroadcast` - Verify message delivery to connected clients.
3. `TestServerReplay` - Verify `Last-Event-ID` replay works.
4. `TestHubRegisterUnregister` - Verify client connection management.
5. `TestChannelProviderDefault` - Verify nil provider returns default error.
6. `TestChannelProviderCustom` - Verify custom provider resolves channels.

### Client Tests (`client_test.go`) - Run with WASM + wasmbrowsertest
1. `TestNewClient` - Verify `New().Client()` returns valid `*SSEClient`.
2. `TestClientConnect` - Verify EventSource is created with correct URL.
3. `TestClientOnMessage` - Verify handler receives `SSEMessage` with raw data.
4. `TestClientReconnect` - Verify exponential backoff logic.
5. `TestClientClose` - Verify EventSource is closed and cleaned up (Q21).

### Integration Tests
1. Full round-trip: Server sends, Client receives raw data.
2. Reconnection with `Last-Event-ID`.

**Run all tests**: `./test.sh`

## API Surface After Refactor

### Exported (Public)
```go
// Types
type Config struct { Log func(args ...any) }
type ServerConfig struct { ... }
type ClientConfig struct { ... }
type SSEServer struct { ... }
type SSEClient struct { ... }
type SSEMessage struct { ID, Event, Data string }  // Q22: shared, used by both sides

// Interfaces (server-side only - !wasm)
type ChannelProvider interface {
    ResolveChannels(r *http.Request) ([]string, error)
}
type SSEPublisher interface {
    Publish(data string, channels []string)
    PublishEvent(event, data string, channels []string)
}

// Functions
func New(*Config) *tinySSE  // Returns private type, but methods are accessible

// SSEServer methods (implements SSEPublisher)
func (*SSEServer) ServeHTTP(w, r)
func (*SSEServer) Publish(data string, channels []string)      // ID auto-generated
func (*SSEServer) PublishEvent(event, data string, channels []string)

// SSEClient methods
func (*SSEClient) Connect()
func (*SSEClient) Close()                    // Q21: Added
func (*SSEClient) OnMessage(func(*SSEMessage))
func (*SSEClient) OnError(func(error))       // Uses tinystring.Err() (Q23)
```

### Unexported (Private)
```go
type tinySSE struct { config *Config }
type hub struct { ... }
type clientConnection struct { ... }  // Renamed from SSEClient in hub.go ✅
```

## Decision References
- Q14: `interfaces_server.go` with `!wasm` build tag
- Q21: `SSEClient.Close()` method for cleanup
- Q22: `models.go` stays shared (SSEMessage used by both client and server)
- Q23: Use `tinystring.Err()` instead of `SSEError` struct

