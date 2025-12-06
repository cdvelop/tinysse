# Architecture: Client Decoupling

## Context
Currently, the WASM client automatically assumes the incoming data is JSON and attempts to Unmarshal it into a specific struct.

**Problems**:
1.  **Opinionated**: Forces JSON usage.
2.  **Performance**: Double decoding if the user wants to handle raw bytes or use a different decoder (e.g., tinyjson).
3.  **SRP Violation**: The transport layer should not be a parser.

## Design Principle
**tinysse is a pure SSE transport layer**. It should:
- Deliver raw SSE events (id, event type, data) to the consumer.
- NOT parse, decode, or assume any data format.
- NOT handle authentication, user management, or routing.

## Proposed Design

### Standard SSE Format (Server sends)
The server sends **standard SSE format**, NOT JSON:
```
id: 123
event: update
data: {"user": "john", "action": "login"}

```
Note: The `data` field contains raw bytes. It could be JSON, plain text, or any format.

**Event field behavior**: If `Event` is empty, the `event:` line is omitted (browser defaults to "message" type). This keeps the stream compact.

### SSEMessage Struct
```go
// models.go (shared, no build tag)
type SSEMessage struct {
    ID    string // SSE "id:" field - Required. Used for Last-Event-ID reconnection.
    Event string // SSE "event:" field - Optional. Allows routing to different handlers.
    Data  []byte // SSE "data:" field - RAW bytes, library does NOT parse.
}
```

**Why `Event` field?**
The SSE protocol defines an optional `event:` field that allows the server to categorize messages:
```
id: 123
event: user_created
data: {"id": 1, "name": "John"}

id: 124
event: notification
data: New message received
```

The client can then route messages based on `Event`:
```go
client.OnMessage(func(msg *tinysse.SSEMessage) {
    switch msg.Event {
    case "user_created":
        handleUserCreated(msg.Data)
    case "notification":
        showNotification(string(msg.Data))
    case "", "message":  // Default event type
        handleGenericMessage(msg.Data)
    }
})
```

If not used, `Event` will be empty string (browser treats as "message" type).

### Handler Pattern
The library passes raw `SSEMessage` to a user-defined handler:

```go
// SSEClient exposes a method to set the handler
type SSEClient struct {
    *tinySSE
    config  *ClientConfig
    handler func(msg *SSEMessage)
    // ...
}

// OnMessage sets the handler for incoming messages
func (c *SSEClient) OnMessage(handler func(msg *SSEMessage)) {
    c.handler = handler
}

// OnError sets the handler for errors
func (c *SSEClient) OnError(handler func(err error)) {
    c.errorHandler = handler
}
```

### Usage Example
```go
client := tinysse.New(&tinysse.Config{
    Log: console.Log,
}).Client(&tinysse.ClientConfig{
    Endpoint: "/events",
})

client.OnMessage(func(msg *tinysse.SSEMessage) {
    // User decides how to decode msg.Data
    switch msg.Event {
    case "update":
        var data MyStruct
        tinyjson.Unmarshal(msg.Data, &data)
        controller.HandleUpdate(data)
    case "notification":
        showNotification(string(msg.Data))
    }
})

client.OnError(func(err error) {
    console.Error("SSE error:", err)
})

client.Connect()
```

## Implementation Steps
1.  Remove `encoding/json` import from client.go (WASM).
2.  Simplify `SSEMessage` - remove `HandlerID`, keep only SSE standard fields.
3.  Parse SSE text format in JS callback, NOT JSON.
4.  Add `OnMessage()` and `OnError()` methods to `SSEClient`.
5.  Server uses standard SSE text format for sending.

## Benefits
- **Zero parsing overhead** in the library.
- **Full flexibility** for the consumer to use any decoder.
- **Smaller WASM binary** (no `encoding/json` import).
- **True SRP**: Transport only, no business logic.
