# Refactor TinySSE Architecture

## Objective
Refactor `tinysse` to achieve:
1. **Strict separation of concerns** - SSE transport only, no auth/user management.
2. **Reduce WASM binary size** - Remove `encoding/json` and server-only code.
3. **Decouple data handling** - Pass raw bytes to user-defined handlers.
4. **Clean API** - `New(*Config).Server(*ServerConfig)` / `Client(*ClientConfig)`.

## Design Principles
- **SSE Transport Only**: tinysse handles connection, reconnection, message delivery.
- **No Authentication**: Security is the consumer's responsibility (e.g., crudp/session).
- **No User Management**: Client routing/identification is external. tinysse only knows `clientID` and `channels[]`.
- **Raw Data Delivery**: Library delivers `[]byte`, user parses.
- **Standard SSE Format**: Server sends `id:`, `event:`, `data:` fields (NOT JSON envelope).

## Decisions Made
- **Q1: tinystring dependency** → Keep it. Will be loaded anyway.
- **Q2: SSEMessage.Event field** → Keep it. Allows message categorization per SSE spec.
- **Q3: User management** → Remove completely. External package (crudp) handles it.
- **Q4: crudp integration** → See [ARCH_CRUDP_INTEGRATION.md](./issues/ARCH_CRUDP_INTEGRATION.md).
- **Q5: Channel resolution** → `ChannelProvider` interface (similar to `UserProvider`).
- **Q6: Client identification** → One unique `clientID` per SSE connection (tab).
- **Q7: Authentication** → Cookies (works with SPA/PWA).
- **Q8: SSEServer mounting** → Implements `http.Handler`, consumer mounts where needed.
- **Q9: Event field behavior** → Omit `event:` line if empty (browser defaults to "message").
- **Q10: Default ChannelProvider** → If nil, use default that returns "channel provider not configured" error.
- **Q11: Client/Channels logic** → 1 client = 1 `clientID` with N channels. Same user in multiple tabs = multiple clientIDs, each with same channels.
- **Q12: Testing** → Real tests using `test.sh` (stdlib + WASM with wasmbrowsertest).
- **Q13: Newlines in data** → Hybrid (Option D). tinysse splits by `\n` and sends multiple `data:` lines. Browser reconstructs automatically. No WASM code added.
- **Q14: ChannelProvider location** → `interfaces_server.go` with `//go:build !wasm` tag.
- **Q15: Broadcasting method names** → `Publish()` and `PublishEvent()`.
- **Q16: ClientID exposure** → Internal only, not exposed to client.
- **Q17: SSE ID generation** → Internal correlative using tinystring (no strconv).
- **Q18: crudp access to SSEServer** → Via `SSEPublisher` interface injected into crudp.
- **Q19: HandlerID in SSE** → crudp responsibility. tinysse doesn't know about it. Include in Data or Event.
- **Q20: Carriage return** → Also handled by Option D split (handles `\r\n` and `\n`).
- **Q21: Client Close()** → Yes, SSEClient has `Close()` method to disconnect EventSource.
- **Q22: Models structure** → `models.go` shared (SSEMessage only). No separate server/client models needed.
- **Q23: Error handling** → Use tinystring `Err()` and `Errf()` instead of standard errors package.

## Execution Order
Execute the following documents in order. Update this file as steps are completed.

1.  **[Config Separation](./issues/ARCH_CONFIG_SEPARATION.md)**
    *   Goal: Split monolithic `Config` into `Config` (shared), `ServerConfig`, and `ClientConfig`.
    *   Pattern: `New(*Config).Server(*ServerConfig)` / `Client(*ClientConfig)`.
    *   Make `tinySSE` private, `SSEServer` and `SSEClient` public.

2.  **[Client Decoupling](./issues/ARCH_CLIENT_DECOUPLING.md)**
    *   Goal: Remove JSON decoding from the SSE client.
    *   Pattern: `OnMessage(func(*SSEMessage))` with raw `Data []byte`.
    *   Use standard SSE text format, not JSON envelope.

3.  **[CRUDP Integration](./issues/ARCH_CRUDP_INTEGRATION.md)**
    *   Goal: Define how tinysse integrates with crudp for user/session management.
    *   Pattern: SSEBroadcaster interface, external channel registration.

4.  **[Cleanup & Verification](./issues/ARCH_CLEANUP.md)**
    *   Goal: Ensure no dead code in WASM and all tests pass.
    *   Rename `SSEClient` in hub.go to `clientConnection`.
    *   Delete `hub_wasm.go`.
    *   Verify WASM binary size reduction.

## Status Log
- [ ] 1. Config Separation
- [ ] 2. Client Decoupling
- [ ] 3. CRUDP Integration
- [ ] 4. Cleanup & Verification
