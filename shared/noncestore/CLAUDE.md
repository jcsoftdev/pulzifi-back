# NonceStore Package (`shared/noncestore/`)

In-memory nonce store with 30-second TTL for cross-subdomain token exchange.

## Files

- `store.go` — Nonce store implementation
- `store_test.go` — 6 test functions

## Exported API

### Types
- `NonceEntry` — Stored data: `AccessToken`, `RefreshToken`, `ExpiresIn`, `CreatedAt`
- `Store` — Thread-safe in-memory store (`sync.Mutex`, `map[string]NonceEntry`)

### Functions
- `New() *Store` — Creates empty nonce store

### Methods (`*Store`)
- `Save(nonce, entry)` — Stores entry with current timestamp, lazily prunes expired entries
- `Consume(nonce) *NonceEntry` — One-time retrieval: returns entry and deletes it. Returns nil if expired or not found.
- `Peek(nonce) *NonceEntry` — Read without consuming. Returns nil if expired. Deletes expired entries on access.

## Usage

Used by `shared/bff/` for the cross-subdomain auth flow:
1. User logs in on base domain -> BFF generates nonce, stores tokens
2. User redirected to tenant subdomain with nonce in URL
3. Tenant subdomain calls `/callback` -> BFF consumes nonce, sets cookies
4. Base domain calls `/set-base-session` -> BFF peeks nonce (does not consume), sets base cookies

## Notes

- 30-second TTL — nonces expire quickly for security
- `Save` lazily prunes all expired entries (garbage collection on write)
- `Consume` is one-time: entry is deleted after retrieval
- `Peek` is non-destructive: entry remains available for subsequent reads

## Architecture Improvements

### Redis-Backed Nonce Store
The in-memory store is **node-local** — if the login request hits instance A and the callback hits instance B, the nonce will not be found. For multi-instance deployments:
1. Replace in-memory map with Redis `SETEX` (automatic 30s TTL, no manual pruning needed)
2. Use Redis `GETDEL` for `Consume` (atomic get-and-delete)
3. Use Redis `GET` for `Peek`
4. Define a `NonceStore` interface so both implementations are swappable
5. Wire via config flag in `cmd/server/main.go`

### Security Hardening
- Add maximum nonce store size to prevent memory exhaustion from abuse (limit concurrent pending auth flows)
- Consider adding rate limiting on nonce creation (tied to IP or session)
- Log nonce consumption failures for security auditing
