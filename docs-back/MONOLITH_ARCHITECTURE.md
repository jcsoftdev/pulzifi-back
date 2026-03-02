# Monolith Architecture

## Overview

Pulzifi Backend runs as a **unified monolith** — a single Go binary serving all HTTP routes, gRPC services, and background workers. Go on port 3000 is the sole entry point for all traffic.

**Stack:** Go 1.25 + Chi router + PostgreSQL (schema-per-tenant) + Redis (optional) + Next.js frontend

---

## Entry Points

| Binary | Location | Purpose |
|--------|----------|---------|
| **Server** | `cmd/server/` | HTTP (:3000) + gRPC (:9000) + reverse proxy to Next.js |
| **Worker** | `cmd/worker/` | Background monitoring checks (can also run in-process) |
| **Migrate** | `cmd/migrate/` | Database migration runner |

The server can run in two modes controlled by `ENABLE_WORKERS`:
- **All-in-One** (default): API + background workers in a single process
- **API-only** (`ENABLE_WORKERS=false`): API only, workers run separately via `cmd/worker/`

---

## HTTP Routing

Go on port 3000 handles all incoming traffic:

| Route Pattern | Handler | Middleware |
|---------------|---------|-----------|
| `/api/auth/*` | BFF handler (`shared/bff/handler.go`) | CORS, Rate limiter |
| `/api/v1/*` | Module routes | CORS, Rate limiter, Tenant, Logging |
| `/health` | Health check | None |
| `/api/v1/swagger/*` | Swagger UI | Tenant (bypassed by public path) |
| `/*` | Reverse proxy to Next.js (:3001) | None |

### Middleware Stack for `/api/v1/*`

1. CORS (wildcard subdomain support)
2. Rate limiter
3. Tenant middleware (subdomain → PostgreSQL schema)
4. Response logger
5. Request logger
6. Module-specific middleware (auth, org membership, etc.)

---

## Module Registry

All modules implement the `ModuleRegisterer` interface:

```go
type ModuleRegisterer interface {
    RegisterHTTPRoutes(router chi.Router)
    ModuleName() string
}
```

Registration happens in `cmd/server/modules.go` via `registerAllModulesInternal()`.

### Registered Modules (17 total)

| Module | Responsibility |
|--------|---------------|
| **Auth** | User authentication, JWT, OAuth2 (Google, GitHub), password reset |
| **Admin** | User registration approval workflow |
| **Email** | Email template rendering and delivery (Resend) |
| **Organization** | Org lifecycle, subdomain provisioning, membership |
| **Workspace** | Workspace CRUD, member roles (Owner/Editor/Viewer) |
| **Page** | URL registration for monitoring within workspaces |
| **Alert** | Alert creation and management on changes |
| **Monitoring** | Check scheduling, frequency config, background workers |
| **Integration** | Third-party webhooks (Slack, Discord, custom) |
| **Insight** | LLM-powered analysis of detected changes (OpenRouter) |
| **Report** | Report generation from monitoring data |
| **Usage** | Resource usage tracking against plan limits |
| **Dashboard** | Aggregated org-wide statistics |
| **Team** | Organization-level member invitations |
| **Snapshot** | Playwright capture, HTML/screenshot storage, change detection |

Additionally, `modules/infra/extractor/` is a Node.js Playwright service (runs separately on :3005).

---

## Module Structure

Every Go module follows hexagonal architecture:

```
modules/{name}/
├── domain/
│   ├── entities/         # Business models (no external imports)
│   ├── repositories/     # Interface definitions only
│   ├── services/         # Shared domain logic
│   ├── errors/           # Business exceptions
│   └── value_objects/    # Immutable value types
├── application/
│   └── {use_case}/       # One directory per use case
│       ├── handler.go    # Orchestration logic
│       ├── request.go    # Input DTO
│       ├── response.go   # Output DTO
│       └── handler_test.go
└── infrastructure/
    ├── http/             # REST routes and HTTP handlers
    │   └── module.go     # ModuleRegisterer implementation
    ├── grpc/             # gRPC server/client + .proto files
    ├── persistence/      # PostgreSQL + in-memory (test) repos
    └── messaging/        # Event publishing
```

**Dependency rule (strictly enforced):** `domain` <- `application` <- `infrastructure`. No imports between modules.

---

## Shared Packages

All in `shared/` — technical utilities only, no business logic:

| Package | Purpose |
|---------|---------|
| `config/` | Env var loading with defaults |
| `database/` | PostgreSQL connection pool + migration runner |
| `middleware/` | Tenant extraction, JWT validation, rate limiter, org membership |
| `bff/` | BFF auth routes (login, callback, refresh, logout, set-base-session) |
| `noncestore/` | In-memory nonce store (30s TTL) for cross-subdomain token exchange |
| `pubsub/` | SSE brokers (CheckBroker, InsightBroker) |
| `router/` | Module registration registry |
| `static/` | Reverse proxy to Next.js / static file serving |
| `swagger/` | Swagger UI setup for Chi |
| `eventbus/` | In-memory pub/sub for async module communication |
| `cache/` | Redis client |
| `ai/` | OpenRouter LLM client |
| `html/` | HTML text extraction |
| `logger/` | Zap structured logging |
| `http/` | Shared HTTP response helpers |

---

## Multi-Tenancy

- Tenant extracted from subdomain: `tenant1.app.com` → resolve via `public.organizations` → set `X-Tenant` context
- `public` schema: users, organizations, sessions, roles, permissions
- `<tenant>` schema: workspaces, pages, checks, insights, alerts, integrations
- All repos call `SET search_path TO <tenant>, public` before queries
- New tenant schema auto-created by `create_tenant_schema()` SQL function on org creation

---

## Authentication (BFF Pattern)

Go handles all cookie management — frontend never touches raw tokens.

**Routes:** `/api/auth/login`, `/api/auth/callback`, `/api/auth/set-base-session`, `/api/auth/refresh`, `/api/auth/logout`

**Cross-subdomain flow:**
1. User logs in at `app.domain.com` → POST `/api/auth/login`
2. BFF generates nonce, stores tokens in NonceStore (30s TTL)
3. Returns nonce + tenant to frontend
4. Frontend redirects to `tenant.domain.com/api/auth/callback?nonce=<uuid>`
5. BFF consumes nonce → sets HttpOnly cookies on tenant subdomain
6. Redirect to `tenant.domain.com/workspaces`

---

## Inter-Module Communication

| Type | Mechanism | Usage |
|------|-----------|-------|
| **Synchronous** | gRPC (port 9000) | Organization service queries from other modules |
| **Asynchronous** | EventBus (in-memory) | Org creation events → schema provisioning |

Proto files live in `modules/{name}/infrastructure/grpc/proto/`. Clients live in the consuming module's `infrastructure/grpc/` directory. Modules never share structs; they deserialize to their own types.

---

## Background Workers

The monitoring module starts background processes when `ENABLE_WORKERS=true`:

1. **Scheduler** polls PostgreSQL every 30s for due monitoring configs
2. **Worker pool** dispatches checks to the Playwright extractor (:3005)
3. **Change detection** via SHA256 hash comparison of extracted HTML
4. On change: create alert, push SSE event, generate LLM insights (OpenRouter)
5. **SSE brokers** push real-time updates to connected clients

---

## Frontend Integration

- **Next.js 16** (App Router) runs on port 3001
- Go reverse proxy (`shared/static/handler.go`) forwards unmatched routes to Next.js
- `FlushInterval: -1` enables SSE and HMR streaming through the proxy
- Frontend packages: `apps/web/` (application), `packages/ui/` (components), `packages/services/` (API clients), `packages/shared-http/` (tenant-aware HTTP factory)

---

## Development

```bash
make dev          # Start full dev stack (Postgres + Extractor + API + Worker with hot reload)
make dev-web      # Start Next.js on :3001
make down         # Stop dev environment
make build        # Build binary to ./bin/api
make swagger      # Regenerate Swagger docs
make migrate      # Run all migrations
```

---

## Environment Variables

See `shared/config/config.go` and `.env.example`. Key variables:

| Variable | Required | Default | Purpose |
|----------|----------|---------|---------|
| `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` | Yes | - | PostgreSQL connection |
| `CORS_ALLOWED_ORIGINS` | Yes | - | Comma-separated allowed origins |
| `EXTRACTOR_URL` | Yes | - | Playwright extractor service URL |
| `JWT_SECRET` | Prod only | `secret` (dev) | JWT signing key |
| `HTTP_PORT` / `PORT` | No | `3000` | HTTP server port |
| `GRPC_PORT` | No | `9000` | gRPC server port |
| `NEXTJS_URL` | No | `http://localhost:3001` | Next.js reverse proxy target |
| `COOKIE_DOMAIN` | No | - | Cookie scope for cross-subdomain auth |
| `FRONTEND_URL` | No | - | Public-facing frontend URL |
| `ENABLE_WORKERS` | No | `true` | Enable background workers in-process |
| `OPENROUTER_API_KEY` | No | - | AI insights (optional) |
| `OPENROUTER_MODEL` | No | `mistralai/mistral-7b-instruct:free` | LLM model |
| `RESEND_API_KEY` | No | - | Email delivery via Resend |
| `REDIS_HOST` | No | - | Redis caching (optional) |
| `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` | No | - | Google OAuth |
| `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET` | No | - | GitHub OAuth |
| `RATE_LIMIT_REQUESTS` | No | `500` | Max requests per window |
| `RATE_LIMIT_WINDOW` | No | `60s` | Rate limit window duration |

---

## Database Migrations

- Public schema: `shared/database/migrations/public/`
- Tenant schema: `shared/database/migrations/tenant/`
- Format: `000001_description.up.sql` / `000001_description.down.sql`
- Tenant migrations apply to all tenant schemas uniformly

```bash
make migrate                     # Run all migrations
make migrate cmd=down            # Rollback
make migrate scope=public cmd=up # Public schema only
make migrate tenant=demo cmd=up  # Specific tenant
```

---

## Key Files

| File | Purpose |
|------|---------|
| `cmd/server/main.go` | HTTP + gRPC server entry point |
| `cmd/server/modules.go` | Module registration and dependency wiring |
| `shared/router/registry.go` | `ModuleRegisterer` interface and registry |
| `shared/bff/handler.go` | BFF authentication handler |
| `shared/noncestore/store.go` | In-memory nonce store for cross-subdomain auth |
| `shared/static/handler.go` | Reverse proxy to Next.js |
| `shared/config/config.go` | All environment variable loading |
| `shared/middleware/tenant.go` | Subdomain → tenant schema resolution |
| `docker-compose.monolith.yml` | Dev stack (PostgreSQL, LocalStack) |
