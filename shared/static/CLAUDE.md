# Static Package (`shared/static/`)

Reverse proxy to Next.js (development) or static file serving (production).

## Files

- `handler.go` — Proxy and static file setup for unmatched routes

## Exported API

### Functions
- `Setup(router chi.Router, frontendURL, staticDir string, logger *zap.Logger)` — Configures frontend serving. Must be called AFTER registering all API routes.

## Modes

### Development (frontendURL set)
Reverse proxy to Next.js for unmatched routes:
- `/api/*` and `/swagger/*` return 404 (API routes only)
- All other routes proxied to Next.js
- Sets `X-Forwarded-For`, `X-Forwarded-Host`, `X-Forwarded-Proto` headers
- Extracts subdomain from `*.localhost`, `*.app.local`, `*.local` hosts and injects `X-Tenant` header
- `FlushInterval = -1` for SSE and HMR streaming support

### Production (staticDir set)
Serves static files from the built frontend directory:
- `/api/*` and `/swagger/*` return 404
- Other routes serve static files
- Falls back to `index.html` for SPA client-side routing when file not found

## Notes

- Contains deprecated `setupProxy` and `setupStatic` functions (older implementations without NotFound handler pattern)
- The `X-Tenant` injection from subdomain is critical for multi-tenant frontend support in dev mode
