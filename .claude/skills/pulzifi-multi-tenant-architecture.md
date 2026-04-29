---
name: pulzifi-multi-tenant-architecture
description: Use when configuring Pulzifi on Dokploy, working with subdomain routing, tenant extraction from headers, BFF auth flows, or debugging multi-tenant schema isolation
---

# Pulzifi Multi-Tenant Architecture

## Overview

Pulzifi is a Go monolith + Next.js frontend with **schema-per-tenant multi-tenancy**. Each organization owns a PostgreSQL schema. Tenants are extracted from subdomain (e.g., `tenant1.pulzifi.com` → schema `tenant1`), with fallback header priority: `X-Tenant` > `X-Forwarded-Host` > `Host`.

**Core principle:** Request context always contains `tenant` (schema name). All repos call `SET search_path TO "<tenant>", public` before queries. BFF auth handler at `/api/auth/*` manages cross-subdomain cookie exchange via nonces.

## Architecture Layers

### Entry Points (`cmd/`)
- **`cmd/server/`** — HTTP :3000 + gRPC :9000. Mounts `/api/auth/*` (BFF, no tenant middleware) → `/api/v1/*` (modules, WITH tenant middleware) → reverse proxy to Next.js :3001
- **`cmd/worker/`** — Standalone background worker (monitoring scheduler, snapshot capture, insights, alerts, emails)
- **`cmd/migrate/`** — Database migration CLI with scope flags (`all`/`public`/`tenant`, optionally `-tenant <schema>`)

### Module Structure
17 modules in `modules/`. 14 registered in monolith at `/api/v1/*`:

**Modules (all have domain/application/infrastructure hexagonal layers):**
auth, admin, email, organization, workspace, page, alert, monitoring, integration, insight, report, usage, dashboard, team

**Special modules (deviations from hexagonal):**
- `api-docs` — Standalone Gin server, not part of monolith
- `infra` — TypeScript/Bun Playwright scraper, separate Dockerfile, HTTP client in Go
- `snapshot` — Worker-only (no HTTP), separate Dockerfile
- `usage` — All logic in module.go, empty domain/application dirs (needs refactoring)
- `organization` — Only module with gRPC server + domain events + EventBus integration

### Shared Packages (`shared/`)
15 packages: ai/, bff/, cache/, config/, database/, eventbus/, html/, http/, logger/, middleware/, noncestore/, pubsub/, router/, static/, swagger/

**Critical for multi-tenancy:** `shared/middleware/`, `shared/bff/`, `shared/config/`

## Subdomain Extraction & Tenant Resolution

### Extraction Priority (line 124-154 of `shared/middleware/tenant.go`)

In order, stops at first match:
1. **`X-Tenant` header** (set by reverse proxy or explicit client)
2. **`X-Forwarded-Host` header** (set by Traefik/reverse proxy)
3. **`Host` header** (original request, always present)

Extracts **first segment before first dot**:
- `tenant1.pulzifi.com` → `tenant1`
- `app.local` → `app`
- `localhost:3000` → `localhost`

### Validation (lines 156-170)

```go
isValidSubdomain(subdomain) → subdomain != "" && subdomain != "localhost"

isGenericDomain(subdomain) → subdomain in ["app", "localhost", "127.0.0.1"]
```

- Generic domains (dev mode): Use subdomain as schema directly (no DB lookup)
- Other subdomains (prod): Query `public.organizations` table

### Schema Resolution (lines 172-177)

```sql
SELECT schema_name FROM public.organizations WHERE subdomain = $1 AND deleted_at IS NULL LIMIT 1
```

If found: Store `schema_name` in context (e.g., `tenant1_schema`)
If not found: Return HTTP 404 "Organization not found"

### Public Paths (bypass tenant middleware)

These paths do NOT require tenant resolution (line 97-121):

```
/swagger, /health, /docs, /auth/login, /auth/register, /auth/callback,
/auth/check-subdomain, /auth/me, /auth/refresh, /auth/logout,
/auth/forgot-password, /auth/reset-password, /auth/oauth, /auth/providers,
/auth/csrf, /admin
```

**Critical:** `/api/auth/callback` is public. Used by BFF auth flow for cross-subdomain nonce exchange.

## BFF Auth Flow (shared/bff/handler.go)

### Routes Mounted at `/api/auth/*` (NO tenant middleware)

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/login` | Authenticates user, generates nonce (30s TTL), sets HttpOnly cookies on current origin |
| POST | `/refresh` | Reads refresh token from cookie, issues new token pair, updates cookies |
| GET/POST | `/logout` | Revokes refresh token, clears auth cookies and tenant_hint |
| GET | `/callback` | **Cross-subdomain:** Consumes nonce, sets cookies on tenant subdomain, redirects |
| GET | `/set-base-session` | Peeks nonce (non-destructive), sets cookies + tenant_hint on base domain |

### Nonce Store (shared/noncestore/store.go)

In-memory, 30-second TTL. After user authenticates on base domain (`app.pulzifi.com/login`):

1. Handler generates UUID nonce → stores `{access_token, refresh_token, expires_in}`
2. Frontend receives nonce in response
3. Frontend navigates to `https://tenant1.pulzifi.com/api/auth/callback?nonce=<uuid>&redirectTo=/`
4. Callback handler **consumes** nonce (deletes from store), sets HttpOnly cookies on `.pulzifi.com` domain
5. Handler redirects to app
6. Browser now has valid cookies for `tenant1.pulzifi.com`

**Why nonce store isn't durable:** In-memory. Restarts lose nonces. For multi-instance deployments, must swap to Redis-backed implementation.

## HTTP Routing & Middleware Order

### Route Hierarchy (cmd/server/main.go)

```
httpRouter (no tenant middleware)
  │
  ├─ /health (direct handler)
  │
  ├─ /api/auth (BFF handler)
  │   ├─ POST /login
  │   ├─ POST /refresh
  │   ├─ GET /callback
  │   └─ ...
  │
  ├─ /api/v1 (v1Router WITH tenant middleware)
  │   ├─ TenantMiddleware (extracts subdomain → resolves schema)
  │   ├─ ResponseLoggerMiddleware
  │   ├─ LoggingMiddleware
  │   ├─ /swagger (Swagger UI)
  │   └─ [Module routes: /auth, /workspace, /page, ...]
  │
  └─ /* (reverse proxy to Next.js :3001)
```

**Order matters:** Tenant middleware MUST come before other middleware on v1Router so context is available to all handlers.

### Reverse Proxy to Next.js (shared/static/handler.go)

Unmatched routes proxied to Next.js. For development/localhost, injects `X-Tenant` header (lines 58-68):

```go
if contains(host, ".localhost") || hasSuffix(host, ".app.local") {
    tenant := parts[0] // first segment
    if valid(tenant) {
        req.Header.Set("X-Tenant", tenant)
    }
}
```

**For production (`pulzifi.com`):** This condition doesn't match. But `X-Forwarded-Host` is sent by Traefik, so tenant extraction still works via priority #2.

## Database & Schema Isolation

### Public Schema (shared across all tenants)

`public.organizations`, `public.users`, `public.sessions`, `public.roles`, `public.permissions`, `public.plans`, `public.oauth_providers`, `public.invitation_statuses`

12 migrations in `shared/database/migrations/public/`

### Tenant Schemas (one per organization)

Schemas are created by `ProvisionTenantSchema()` in `shared/database/migrator.go`. Each schema contains:

`workspaces`, `pages`, `checks`, `insights`, `monitoring_configs`, `usage_tracking`, `monitored_sections`, `section_rects`, `parent_check_relationships`

12 migrations in `shared/database/migrations/tenant/` applied to all tenant schemas uniformly.

### Query Pattern

Every tenant-aware repo:

```go
type WorkspaceRepository struct {
    db     *sql.DB
    tenant string
}

func (r *WorkspaceRepository) List(ctx context.Context) ([]Workspace, error) {
    // Set search_path for this query
    if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
        return nil, err
    }
    
    // Now query against tenant schema
    rows, err := r.db.QueryContext(ctx, "SELECT * FROM workspaces")
    // ...
}
```

**Validation:** `middleware.GetSetSearchPathSQL(tenant)` rejects invalid schema names (regex: `^[a-zA-Z_][a-zA-Z0-9_]*$`), returns safe no-op if invalid.

## Configuration (shared/config/config.go)

### Critical for Multi-Tenancy

| Variable | Example | Purpose | Required |
|----------|---------|---------|----------|
| `COOKIE_DOMAIN` | `.pulzifi.com` | Cross-subdomain cookie scope. **MUST start with dot** (`.domain` matches `sub.domain`) | **YES** |
| `FRONTEND_URL` | `https://app.pulzifi.com` | Base frontend URL (redirects, BFF nonce exchange) | **YES** |
| `NEXTJS_URL` | `http://nextjs:3001` (Docker) | Internal URL to Next.js service | **YES** |
| `CORS_ALLOWED_ORIGINS` | `https://app.pulzifi.com,https://*.pulzifi.com` | Wildcard allows all subdomains | **YES** |
| `JWT_SECRET` | (random 32+ bytes, prod only) | Signs JWT tokens | **YES** (prod) |
| `HTTP_PORT` | `3000` | Go server port | No (default 3000) |
| `GRPC_PORT` | `9000` | gRPC port | No |
| `EXTRACTOR_URL` | `http://extractor:3005` | Internal Playwright scraper | **YES** |
| `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` | (`postgres:5432`) | PostgreSQL connection | **YES** |
| `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD` | (`redis:6379`) | Optional. Graceful degradation if missing | No |
| `ENABLE_WORKERS` | `true`/`false` | Run monitoring/snapshot/insights in monolith (default true) | No |

## Dokploy Deployment

### Architecture

Dokploy orchestrates 5+ services:

| Service | Image | Port | Notes |
|---------|-------|------|-------|
| **api** | `Dockerfile.api` | 3000 | Production Go monolith |
| **worker** | `Dockerfile.worker` | N/A | Background jobs (separate from API) |
| **scraper** | `modules/infra/scraper/Dockerfile` | 3005 | Playwright HTTP service |
| **frontend** | `frontend/Dockerfile` | 3001 | Next.js (internal, proxied by Go) |
| **minio** | Official image | 9000 | Object storage (S3 emulation) |
| **postgres** | Built-in service | 5432 | Database |
| **redis** | Built-in service | 6379 | Cache (optional) |

### DNS Setup

Dokploy uses Traefik for SSL/routing.

**Wildcard required:**
- DNS A record: `*.pulzifi.com` → Dokploy IP
- Or individual subdomains: `app.pulzifi.com`, `tenant1.pulzifi.com`, etc.

Traefik automatically:
- Routes all subdomains to Go container (port 3000)
- Terminates SSL
- Forwards `X-Forwarded-Proto`, `X-Forwarded-Host` headers (so tenant extraction works)

### ENV Setup in Dokploy

When creating the `api` app in Dokploy UI, set:

```
COOKIE_DOMAIN=.pulzifi.com
FRONTEND_URL=https://app.pulzifi.com
NEXTJS_URL=http://nextjs:3001
EXTRACTOR_URL=http://scraper:3005
CORS_ALLOWED_ORIGINS=https://app.pulzifi.com,https://*.pulzifi.com
JWT_SECRET=<generate-32-bytes>
DB_HOST=postgres
DB_PORT=5432
DB_NAME=pulzifi
DB_USER=postgres
DB_PASSWORD=<docker-secret>
REDIS_HOST=redis
REDIS_PORT=6379
HTTP_PORT=3000
GRPC_PORT=9000
ENABLE_WORKERS=true
LOG_LEVEL=info
ENVIRONMENT=production
```

### Network Communication

All services in Dokploy are on the same internal network. Use service names as hostnames:
- Go connects to `postgres:5432` (not localhost)
- Go connects to `redis:6379`
- Go connects to `scraper:3005`
- Next.js container at `nextjs:3001` (Go reverse proxies to this)

### Missing from Dokploy Config

- **Rate limiter is node-local:** If you scale API to 2+ instances, each gets independent token buckets. Client can multiply rate limit by instance count. Future: swap to Redis-backed implementation.
- **Nonce store is in-memory:** Scales to 1 API instance only. Multi-instance: swap to Redis.
- **EventBus is in-memory:** No durability across restarts. For production: implement Kafka adapter.

## Common Mistakes

### ❌ COOKIE_DOMAIN Without Leading Dot
```
COOKIE_DOMAIN=pulzifi.com  # WRONG - cookies only work on exact domain
COOKIE_DOMAIN=.pulzifi.com # RIGHT - matches *.pulzifi.com
```

### ❌ Tenant Extraction Fails in Dokploy
**Symptom:** "Subdomain is required" error
**Cause:** Traefik not sending `X-Forwarded-Host`, or DNS not routing wildcard
**Fix:** Verify Dokploy DNS (wildcard A record) and Traefik config

### ❌ Nonce Expired on `/api/auth/callback`
**Symptom:** "SessionExpired" redirect to /login?error
**Cause:** 30-second nonce TTL exceeded between login and callback
**Fix:** For slow networks, increase nonce TTL in `shared/noncestore/store.go` (line ~30)

### ❌ CORS Error on Wildcard Subdomains
**Symptom:** "Access-Control-Allow-Origin" mismatch
**Cause:** CORS origins don't include wildcard
**Fix:** Set `CORS_ALLOWED_ORIGINS=https://app.pulzifi.com,https://*.pulzifi.com` (Dokploy validates against Origin header from request)

### ❌ Schema Not Found for Tenant
**Symptom:** HTTP 404 "Organization not found"
**Cause:** Subdomain extracted correctly, but no row in `public.organizations` with that subdomain
**Fix:** Ensure organization is created and `subdomain` column matches the URL subdomain

## When Subdomain Extraction Works & Doesn't

| Scenario | Subdomain Extracted | Why |
|----------|-------------------|-----|
| `localhost:3000` | `localhost` | Matches Host header |
| `app.local` | `app` | Matches Host header |
| `tenant1.pulzifi.com` (Dokploy) | `tenant1` | Matches X-Forwarded-Host (from Traefik) |
| `localhost` (no subdomain) | empty | No dot, no match. 400 error if not generic domain. |
| `app.pulzifi.com` (Dokploy) | `app` | `app` is generic domain → uses as schema directly |
| SSR request from Next.js to Go | Depends on headers | Must include X-Tenant or X-Forwarded-Host in request |

## Hexagonal Module Template

Standard structure (most modules follow):

```
modules/{name}/
├── domain/
│   ├── entities/         # Business models, no external imports
│   ├── repositories/     # Interface definitions (no impl)
│   ├── services/         # Domain logic (interfaces + logic)
│   ├── errors/           # Business exceptions
│   └── value_objects/    # Immutable typed values
├── application/
│   └── {use_case}/       # One dir per use case
│       ├── handler.go    # Orchestration
│       ├── request.go    # Input DTO
│       ├── response.go   # Output DTO
│       └── handler_test.go
└── infrastructure/
    ├── http/             # REST routes, module.go registers ModuleRegisterer
    ├── persistence/      # PostgreSQL + in-memory test impl
    ├── grpc/             # (only organization module)
    ├── messaging/        # Event publishing
    └── {feature}/        # Email, OAuth, schedulers, etc.
```

**Dependency rule:** `domain` ← `application` ← `infrastructure`. Never import between modules; use interfaces.

## Debugging Tips

### Log Tenant Extraction
Enable `LOG_LEVEL=debug` to see tenant resolution:
```
DEBUG: Tenant resolved subdomain=tenant1 schema=tenant1_schema path=/api/v1/workspaces
```

### Verify Schema Exists
```sql
SELECT schema_name FROM public.organizations WHERE subdomain = 'tenant1';
\dn  -- list all schemas
```

### Check Cookie Domain
Open DevTools → Application → Cookies. The `access_token` cookie should have `Domain: .pulzifi.com` (not `pulzifi.com`).

### Trace Subdomain Extraction
`shared/middleware/tenant.go` lines 124-154 show the priority. Add logging if extraction fails:
```go
logger.Debug("Failed to extract subdomain",
    zap.String("host", r.Host),
    zap.String("x-forwarded-host", r.Header.Get("X-Forwarded-Host")),
    zap.String("x-tenant", r.Header.Get("X-Tenant")))
```

## Real-World Impact

**Before:** Monolith per tenant (5 deployments, 5x cost)
**After:** Schema-per-tenant (1 deployment, 1x cost, dynamic scaling)

**Architectural benefit:** Code doesn't change when adding tenants. `SET search_path` automatically isolates queries. No tenant ID leakage across schemas.

**Production concern:** Rate limiter + nonce store + EventBus are node-local. Single instance only (or swap implementations). Plan to upgrade before horizontal scaling.

## Initial Setup: Creating First Organization

After deploying to Dokploy with empty database:

```sql
-- 1. Run migrations
go run ./cmd/migrate -cmd up

-- 2. Create first organization (substitute with your values)
INSERT INTO public.organizations 
  (id, subdomain, schema_name, name, deleted_at) 
VALUES 
  (gen_random_uuid(), 'app', 'public', 'Main Organization', NULL);

-- 3. Create first user (hash password with bcrypt, e.g., "password" → $2a$10/...)
INSERT INTO public.users 
  (id, email, password_hash, is_super_admin, created_at) 
VALUES 
  (gen_random_uuid(), 'admin@pulzifi.com', '<BCRYPT_HASH>', true, now());

-- 4. Create additional tenant (if needed)
INSERT INTO public.organizations 
  (id, subdomain, schema_name, name, deleted_at) 
VALUES 
  (gen_random_uuid(), 'tenant1', 'tenant1_schema', 'Tenant 1', NULL);

-- Verify:
SELECT subdomain, schema_name FROM public.organizations;
\dn  -- List all schemas
```

**Why `app.pulzifi.com` uses `public` schema:** The base domain is a special case. All users without a subdomain (or with `app`) run against the shared public schema. Not recommended for production multi-tenant (exposes cross-tenant data). For production, create a separate schema per organization.

## Generating Secure Secrets

For Dokploy env vars, generate cryptographically secure values:

```bash
# JWT_SECRET (32 bytes = 256 bits)
openssl rand -hex 32

# Database/Redis passwords (16 bytes)
openssl rand -hex 16

# MinIO/S3 credentials
openssl rand -hex 16  # access key
openssl rand -hex 32  # secret key
```

Store in Dokploy **Application Secrets** (encrypted), not in git.

## Multi-Instance Scaling (Future)

**When adding 2+ API instances:**

1. **Rate limiter:** Replace `sync.Map` with Redis backend (use Redis `INCR` + `EXPIRE`)
2. **Nonce store:** Swap to Redis-backed store (same 30s TTL, but durable)
3. **EventBus:** Implement Kafka adapter (interface already designed for this)
4. **Session affinity:** Optional. Traefik can route by cookie → same instance

Current code uses in-memory implementations. See architecture documentation in `shared/` for interface signatures. All are designed to be swappable.
