# Middleware Package (`shared/middleware/`)

HTTP middleware for tenant extraction, auth, rate limiting, and request logging.

## Files

- `tenant.go` — Tenant extraction from subdomain and schema resolution
- `auth_provider.go` — Global singleton holder for auth middleware
- `organization.go` — Organization membership validation
- `rate_limiter.go` — Token bucket rate limiter per IP
- `logging.go` — Request logging middleware
- `response_logger.go` — Response logging middleware
- `health.go` — Health check handler (legacy Gin, likely unused)
- `tenant_test.go` — 10 test functions with 50+ subtests
- `rate_limiter_test.go` — 7 test functions

## Tenant Middleware (`tenant.go`)

### Exported API
- `TenantMiddleware(db *sql.DB) func(http.Handler) http.Handler` — Extracts subdomain, resolves to `schema_name` via `public.organizations` table, stores in context
- `GetTenantFromContext(ctx) string` — Get tenant schema from context
- `GetSubdomainFromContext(ctx) string` — Get subdomain from context
- `GetSetSearchPathSQL(tenant) string` — Returns `SET search_path TO "<tenant>", public` with SQL injection prevention (returns `SELECT 1` for invalid names)
- `GetTenantFromContextOrError(ctx) (string, error)` — Returns error if tenant missing
- `RequireTenant(next) http.Handler` — Returns 400 if no tenant in context

### Subdomain Extraction Priority
1. `X-Tenant` header
2. `X-Forwarded-Host` header
3. `Host` header

### Public Paths (bypass tenant resolution)
`/swagger`, `/health`, `/docs`, `/auth/login`, `/auth/register`, `/auth/check-subdomain`, `/auth/me`, `/auth/refresh`, `/auth/logout`, `/auth/forgot-password`, `/auth/reset-password`, `/auth/oauth`, `/auth/providers`, `/auth/csrf`, `/admin`

## Auth Provider (`auth_provider.go`)

- `AuthMiddleware` — Global auth middleware singleton
- `OrgMiddleware` — Global organization middleware singleton
- `SetAuthMiddleware(middleware)` — Set at server startup
- `SetOrganizationMiddleware(middleware)` — Set at server startup

## Organization Middleware (`organization.go`)

- `OrganizationMiddleware` — Validates user belongs to org by querying `public.organization_members`
- `RequireOrganizationMembership(next) http.Handler` — Returns 403 if not a member

## Rate Limiter (`rate_limiter.go`)

- `RateLimiter` — In-memory token bucket per IP with `sync.Map`
- `NewRateLimiter(maxTokens, window) *RateLimiter` — Creates limiter, starts background cleanup
- `Handler(next) http.Handler` — Returns 429 with `Retry-After` header when exhausted
- `Stop()` — Terminates cleanup goroutine

### IP Extraction Priority
1. `X-Forwarded-For` (first IP)
2. `X-Real-IP`
3. `RemoteAddr`

## Logging (`logging.go`, `response_logger.go`)

- `LoggingMiddleware(next) http.Handler` — Logs requests with method, path, tenant, hostname, status, duration
- `ResponseLoggerMiddleware(next) http.Handler` — Logs responses with status and duration

Both wrap `http.ResponseWriter` with `Unwrap()` for SSE/Flusher compatibility.

## Health Check (`health.go`)

- `HealthCheck() gin.HandlerFunc` — Legacy Gin handler, returns `{"status": "ok"}`. Likely unused in current Chi-based architecture.

## Tests

- `tenant_test.go` — SQL injection prevention, public path classification, context extraction, subdomain extraction, schema validation
- `rate_limiter_test.go` — Limit enforcement, per-IP isolation, window reset, IP extraction

## Architecture Improvements

### Remove or Port `health.go`
`health.go` uses Gin (`gin.HandlerFunc`) but the entire application uses Chi. This file is likely unused since the monolith's health endpoint is registered directly in `cmd/server/main.go` as a Chi handler. Should be removed or ported to `func(w http.ResponseWriter, r *http.Request)`.

### Redis-Backed Rate Limiter
The current rate limiter uses `sync.Map` which is **node-local**. In a multi-instance deployment, each instance maintains its own token buckets, so a client can effectively multiply its rate limit by the number of instances. Replace with:
1. Redis-based token bucket using `INCR`/`EXPIRE` or the Redis Cell module (`CL.THROTTLE`)
2. Keep the same `Handler(next) http.Handler` signature — only the backing store changes
3. Add a config flag (e.g., `RATE_LIMITER_BACKEND=redis|memory`) for development flexibility

### Tenant Schema Validation
`GetSetSearchPathSQL` does basic SQL injection prevention by checking for alphanumeric + underscore characters. Consider additionally:
- Maintaining a cached allowlist of valid tenant schemas (refreshed periodically from `public.organizations`)
- Adding metrics/alerting for rejected schema names (potential attack indicator)

### Public Path Maintenance
The hardcoded `publicPaths` list in `tenant.go` must be updated whenever new unauthenticated routes are added. Consider:
- Moving to a configuration-driven approach
- Or using route registration metadata to automatically determine which paths bypass tenant resolution
