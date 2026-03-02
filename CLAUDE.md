# CLAUDE.md

Instructions for Claude Code when working in this repository.

## Commands

### Development
```bash
make dev          # Start full dev stack (postgres + extractor + API + worker with hot reload)
make dev-web      # Start Next.js on :3001 (Go on :3000 proxies unmatched routes)
make down         # Stop dev environment and remove volumes
make logs service=monolith  # View logs for a specific service
```

### Build & Docs
```bash
make build        # Build API binary to ./bin/api
make swagger      # Regenerate Swagger docs from cmd/server/main.go
```

### Database
```bash
make migrate                     # Run all migrations (public + tenant schemas)
make migrate cmd=down            # Rollback migrations
make migrate cmd=version         # Check current migration version
make migrate scope=public cmd=up # Public schema only
make migrate tenant=demo cmd=up  # Specific tenant schema
```

### Tests
```bash
go test ./...                        # Run all tests
go test ./modules/workspace/...      # Run tests for a specific module
go test -v -run TestName ./path/...  # Run a single test
go test -race ./...                  # Run with race detector
```

### Frontend (from `frontend/` directory)
```bash
bun dev            # Start Next.js dev server on :3001
bun run build      # Production build
bun run lint:fix   # Format and lint with Biome
```

## Architecture

**Backend:** Go 1.25 + Chi router + Hexagonal Architecture + Vertical Slicing + Multi-Tenant by subdomain
**Frontend:** Next.js 16 + React 19 + Tailwind CSS + Bun + Turborepo + Biome
**Entry points:** `cmd/server/` (HTTP :3000 + gRPC :50051), `cmd/worker/` (background jobs), `cmd/migrate/` (DB migrations)

### Module Structure

All 17 modules in `modules/` follow this layout:
```
modules/{name}/
├── domain/
│   ├── entities/         # Business models (no external imports)
│   ├── repositories/     # Interface definitions only
│   ├── services/         # Shared domain logic
│   ├── errors/           # Business exceptions
│   └── value_objects/
├── application/
│   └── {use_case}/       # One directory per use case
│       ├── handler.go    # Orchestration logic
│       ├── request.go    # Input DTO
│       ├── response.go   # Output DTO
│       └── handler_test.go
└── infrastructure/
    ├── http/             # REST routes and HTTP handlers
    ├── grpc/             # gRPC server/client + .proto files
    ├── persistence/      # PostgreSQL + in-memory (test) implementations
    └── messaging/        # Event publishing
```

**Dependency rule (strictly enforced):** `domain` ← `application` ← `infrastructure`. No imports between modules.

### Coding Conventions

- Package naming: directory `create_check` → `package createcheck` (no underscores in package names)
- Tenant-aware repos: constructor accepts `tenant string`, calls `middleware.GetSetSearchPathSQL(tenant)` before queries
- Use `context.WithTimeout` for external calls in goroutines
- In-memory repository implementations for tests (no database dependency in unit tests)
- One use case = one directory under `application/`

### Multi-Tenancy

- Tenant extracted from subdomain: `tenant1.app.com` → `X-Tenant: tenant1` header
- Middleware resolves subdomain → `schema_name` via `public.organizations` table
- All repos call `SET search_path TO <tenant>, public` before queries
- PostgreSQL schema-per-tenant: `public` schema holds users/orgs; tenant schemas hold workspaces/pages/checks/etc.
- New tenant schema auto-created by `create_tenant_schema()` SQL function on org creation

### HTTP Routing

- Go on :3000 is the single entry point
- `/api/auth/*` → BFF handler (`shared/bff/handler.go`) — cookie management, nonce exchange
- `/api/v1/*` → Module routes — tenant middleware extracts schema
- `/*` → Reverse proxy to Next.js on :3001

### Inter-Module Communication

- **Synchronous:** gRPC (proto in `infrastructure/grpc/proto/`, client in `infrastructure/grpc/`)
- **Asynchronous:** EventBus (dev) / Kafka (prod) — events published from `infrastructure/messaging/`
- Never share structs between modules; deserialize to own types

### Shared Packages (`shared/`)

Technical utilities only — no business logic:
- `config/` — env var loading with defaults
- `database/` — PostgreSQL connection pool + migration runner
- `middleware/` — tenant extraction, JWT validation, rate limiter, org membership
- `bff/` — BFF auth routes (login, refresh, logout, callback, set-base-session)
- `noncestore/` — in-memory nonce store (30s TTL) for cross-subdomain token exchange
- `pubsub/` — SSE brokers (CheckBroker, InsightBroker)
- `ai/` — OpenRouter LLM client
- `eventbus/` — in-memory pub/sub
- `logger/` — Zap structured logging
- `html/` — HTML text extraction
- `cache/` — Redis client
- `router/` — module registration registry
- `static/` — reverse proxy to frontend / static file serving

### Database Migrations

- Public schema: `shared/database/migrations/public/` — users, orgs, sessions, roles
- Tenant schema: `shared/database/migrations/tenant/` — workspaces, pages, checks, insights
- Format: `000001_description.up.sql` / `000001_description.down.sql`
- Tenant migrations apply to all tenant schemas uniformly

### Background Jobs

Worker process (`cmd/worker/`) handles:
- Scheduled monitoring checks (30s poll interval)
- Snapshot change detection via Playwright extractor
- AI insight generation after changes (OpenRouter)
- Alert creation and email notifications

### Frontend (`frontend/`)

- `apps/web/` — Next.js App Router application
- `packages/ui/` — Atomic Design component library
- `packages/services/` — API client layer (auth-api, page-api)
- `packages/shared-http/` — Tenant-aware HTTP factory (server/browser clients)
- Route groups: `(public)/`, `(main)/` (authenticated), `(auth)/`
- Features follow vertical slicing: `features/{name}/` with UI, application, domain layers
- Middleware (`proxy.ts`) validates sessions, handles token refresh, redirects to login

### Key Config Variables

See `shared/config/config.go` and `.env.example`. Critical:
- `DB_*` / `DATABASE_URL` — PostgreSQL connection
- `JWT_SECRET` — JWT signing key
- `CORS_ALLOWED_ORIGINS` — Comma-separated allowed origins
- `EXTRACTOR_URL` — Playwright screenshot service
- `OPENROUTER_API_KEY`, `OPENROUTER_MODEL` — AI insights (optional)
- `RESEND_API_KEY` — Email via Resend (optional)
- `COOKIE_DOMAIN` — Cookie scope for cross-subdomain auth
