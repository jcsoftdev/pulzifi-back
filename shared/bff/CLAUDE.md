# BFF Package (`shared/bff/`)

Backend-For-Frontend auth handler for cross-subdomain cookie management.

## Files

- `handler.go` — BFF auth routes mounted at `/api/auth/*`

## Exported API

### Types
- `Handler` — BFF handler composing login/logout/refresh handlers, token service, nonce store, cookie config
- `HandlerDeps` — Dependency injection struct with all required dependencies

### Functions
- `NewHandler(deps HandlerDeps) *Handler` — Creates BFF handler

### Methods (`*Handler`)
- `RegisterRoutes(r chi.Router)` — Mounts all BFF auth routes

## Routes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/login` | Authenticates user, generates nonce for cross-subdomain redirect, sets HttpOnly cookies |
| POST | `/refresh` | Reads refresh_token cookie, refreshes tokens, sets new cookies |
| POST | `/logout` | Revokes refresh token, clears auth + tenant_hint cookies |
| GET | `/logout` | Same as POST but redirects (supports `?redirectTo=`, default `/login`) |
| GET | `/callback` | Consumes nonce, sets HttpOnly cookies on tenant subdomain, redirects |
| GET | `/set-base-session` | Peeks nonce (does NOT consume), sets cookies + tenant_hint on base domain |

## Cookie Management

- Sets `access_token` and `refresh_token` as HttpOnly, Secure cookies
- Sets `tenant_hint` cookie (7-day expiry, HttpOnly, SameSite=Lax) for subdomain routing
- Cookie domain and secure flag configurable via `HandlerDeps`

## Dependencies

- `modules/auth` — login, logout, refresh_token handlers, token service, cookies package
- `shared/noncestore` — Nonce store for cross-subdomain token exchange
