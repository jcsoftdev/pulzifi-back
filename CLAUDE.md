# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Development
```bash
make dev          # Start full dev stack (postgres + extractor + API + worker with hot reload)
make dev-web      # Start Caddy proxy (:3000) + Next.js (:3001) for frontend dev
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

**Backend:** Go + Hexagonal + Vertical Slicing + Multi-Tenant by subdomain
**Entry points:** `cmd/server/` (HTTP :9090 + gRPC :50051), `cmd/worker/` (background jobs), `cmd/migrate/` (DB migrations)

### Module Structure

All modules in `modules/` follow this layout:
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
    └── messaging/        # Kafka publisher/subscriber
```

**Dependency rule (strictly enforced):** `domain` → `application` → `infrastructure`. No imports between modules.

### Multi-Tenancy

- Tenant extracted from subdomain: `tenant1.app.com` → `X-Tenant: tenant1` header
- Middleware sets `context.Value(TenantKey, "tenant1")`
- All repos call `SET search_path TO <tenant>, public` before queries (see `middleware.GetSetSearchPathSQL`)
- PostgreSQL schema-per-tenant: public schema holds users/orgs; tenant schemas hold workspaces/pages/checks/etc.
- New tenant schema auto-created by `create_tenant_schema()` SQL function on org creation

### Inter-Module Communication

- **Synchronous:** gRPC — proto in `infrastructure/grpc/proto/`, client in `infrastructure/grpc/<module>_client.go`
- **Asynchronous:** EventBus (dev) / Kafka (prod) — events published from `infrastructure/messaging/publisher.go`
- Never share structs between modules; deserialize to own types

### Shared Packages (`shared/`)

Technical utilities only — no business logic:
- `shared/config/` — all env vars with defaults
- `shared/database/` — PostgreSQL connection + migrations
- `shared/middleware/` — tenant extractor, JWT validator, rate limiter
- `shared/ai/` — OpenRouter client
- `shared/eventbus/` — in-memory pub/sub
- `shared/logger/` — Zap structured logging

### Database Migrations

- Public schema: `shared/database/migrations/public/` — users, orgs, sessions, roles
- Tenant schema: `shared/database/migrations/tenant/` — workspaces, pages, checks, insights, reports
- Format: `000001_description.up.sql` / `000001_description.down.sql`
- Add new tenant tables to tenant migrations (all tenants share the same schema structure)

### Background Jobs

Runs in `cmd/worker/`:
- Scheduled monitoring checks (Asynq/Redis)
- Snapshot change detection (`modules/snapshot/application/worker.go`)
- AI insight generation after changes (`modules/insight/`)
- Email sending with retry logic

### Key Config Variables

See `shared/config/config.go` and `.env.example` for all options. Critical ones:
- `DB_*` — PostgreSQL connection
- `JWT_SECRET` — JWT signing key
- `OPENROUTER_API_KEY`, `OPENROUTER_MODEL` — AI insights
- `RESEND_API_KEY` — Email (optional, disabled if not set)
- `EXTRACTOR_URL` — Playwright screenshot service

### Frontend (`frontend/`)

Next.js app with Bun + Turbo + Biome. Uses Atomic Design for UI components (`packages/ui/`) and DDD + Vertical Slicing for app features (`apps/web/src/features/`).
