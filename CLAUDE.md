# CLAUDE.md

Instructions for Claude Code when working in this repository.

## Commands

### Development
```bash
make dev          # Start full dev stack (postgres + localstack + extractor + API + worker with hot reload)
make dev-web      # Start Next.js on :3001 (Go on :3000 proxies unmatched routes)
make down         # Stop dev environment and remove volumes
make logs service=monolith  # View logs for a specific service
make clean        # Stop all containers and prune Docker resources
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
bun run format     # Format with Biome
bun run type-check # TypeScript type checking
```

### Scaffolding & Validation Tools
```bash
./tools/scripts/new-module.sh <module-name>              # Scaffold a new hexagonal module with CLAUDE.md
./tools/scripts/new-migration.sh <scope> <description>   # Scaffold a new migration (up + down SQL files)
./tools/scripts/check-architecture.sh                    # Verify hexagonal architecture rules
./tools/scripts/validate-build.sh                        # Pre-commit validation (build, vet, tests, types, arch)
```

## Architecture

**Backend:** Go 1.25 + Chi router + Hexagonal Architecture + Vertical Slicing + Multi-Tenant by subdomain
**Frontend:** Next.js 16 + React 19 + Tailwind CSS + Bun + Turborepo + Biome
**Entry points:** `cmd/server/` (HTTP :3000 + gRPC :9000), `cmd/worker/` (background jobs), `cmd/migrate/` (DB migrations)
**Deployment:** Railway (4 services: api, worker, extractor, frontend)

### Entry Points (`cmd/`)

- **`cmd/server/`** — Main HTTP + gRPC monolith. Loads config, connects to PostgreSQL/Redis, registers 14 module routes under `/api/v1/*`, mounts BFF auth at `/api/auth/*`, proxies unmatched routes to Next.js, starts gRPC server for Organization service. If `ENABLE_WORKERS=true`, also runs background monitoring processes (all-in-one mode).
- **`cmd/worker/`** — Standalone background worker. Runs monitoring scheduler, snapshot change detection, AI insight generation, alert creation, and email notifications.
- **`cmd/migrate/`** — Database migration CLI. Supports flags: `-cmd` (up/down/version/force), `-scope` (all/public/tenant), `-tenant` (specific schema), `-steps`.

### Module Structure

There are 17 module directories in `modules/`. 14 are registered in the monolith server, plus 3 special-purpose modules:

| Module | Type | Description |
|--------|------|-------------|
| admin | API | User registration approval workflow (SUPER_ADMIN) |
| alert | API | Change detection alert notifications |
| api-docs | Standalone | Swagger/OpenAPI documentation hub (Gin, :9000, not part of monolith) |
| auth | API | Authentication, JWT, OAuth (Google/GitHub), sessions, roles |
| dashboard | API | Aggregated organization statistics |
| email | API | Email service with Resend provider and HTML templates |
| infra | Standalone | TypeScript/Bun Playwright scraper service (lives in `infra/scraper/`, not Go) |
| insight | API | AI-powered insight generation via OpenRouter LLM |
| integration | API | Third-party integrations (webhooks, Slack, Teams, Discord) |
| monitoring | API + Worker | Page monitoring scheduler, check execution, notifications |
| organization | API + gRPC | Organization entity, multi-tenant schema management, events |
| page | API | Monitored page CRUD with tags |
| report | API | Monitoring report generation and storage |
| snapshot | Worker | Background snapshot capture, storage (MinIO/Cloudinary), change detection |
| team | API | Organization-level team member management and invitations |
| usage | API | Billing, usage quotas, plan management (SUPER_ADMIN) |
| workspace | API | Workspace CRUD, member management, role-based authorization |

Standard hexagonal module layout (most modules follow this):
```
modules/{name}/
├── domain/
│   ├── entities/         # Business models (no external imports)
│   ├── repositories/     # Interface definitions only
│   ├── services/         # Shared domain logic
│   ├── errors/           # Business exceptions
│   └── value_objects/    # Immutable typed values
├── application/
│   └── {use_case}/       # One directory per use case
│       ├── handler.go    # Orchestration logic
│       ├── request.go    # Input DTO
│       ├── response.go   # Output DTO
│       └── handler_test.go
└── infrastructure/
    ├── http/             # REST routes and HTTP handlers (module.go)
    ├── grpc/             # gRPC server/client + .proto files
    ├── persistence/      # PostgreSQL + in-memory (test) implementations
    ├── messaging/        # Event publishing/subscribing
    └── ...               # Module-specific: ai/, oauth/, cookies/, webhook/, scheduler/, etc.
```

**Notable deviations:**
- `api-docs` — No hexagonal layers; standalone Gin server for Swagger aggregation
- `infra` — TypeScript/Bun service (Playwright/Chromium scraper), lives in `infra/scraper/`, has its own Dockerfile
- `snapshot` — Flat application files (worker service, no HTTP routes), has its own Dockerfile
- `usage` — Domain entities/repositories directories exist but are empty (thin wrapper)
- `monitoring` — Has `application/orchestrator/` and `application/workers/` for background job coordination
- `organization` — Only module with full gRPC implementation (.proto + generated code + server + interceptors), domain events, and event messaging (publisher/subscriber)

**Dependency rule (strictly enforced):** `domain` <- `application` <- `infrastructure`. No imports between modules.

### Coding Conventions

- Package naming: directory `create_check` -> `package createcheck` (no underscores in package names)
- Tenant-aware repos: constructor accepts `tenant string`, calls `middleware.GetSetSearchPathSQL(tenant)` before queries
- Use `context.WithTimeout` for external calls in goroutines
- In-memory repository implementations for tests (no database dependency in unit tests)
- One use case = one directory under `application/`
- Mock files live in `domain/repositories/mocks/` and `domain/services/mocks/`
- HTTP responses use `shared/http` helpers: `RespondJSON`, `RespondError`, `RespondOK`, etc.
- Each module's HTTP layer registers routes via `ModuleRegisterer` interface (`RegisterHTTPRoutes`, `ModuleName`)

### Multi-Tenancy

- Tenant extracted from subdomain: `tenant1.app.com` -> `X-Tenant: tenant1` header
- Middleware resolves subdomain -> `schema_name` via `public.organizations` table
- All repos call `SET search_path TO <tenant>, public` before queries (via `middleware.GetSetSearchPathSQL`)
- PostgreSQL schema-per-tenant: `public` schema holds users/orgs/sessions/roles; tenant schemas hold workspaces/pages/checks/etc.
- New tenant schema auto-created by `ProvisionTenantSchema()` in `shared/database/migrator.go`
- Subdomain extraction priority: `X-Tenant` header > `X-Forwarded-Host` > `Host`

### HTTP Routing

- Go on :3000 is the single entry point
- `/api/auth/*` -> BFF handler (`shared/bff/handler.go`) — cookie management, nonce exchange
- `/api/v1/*` -> Module routes — tenant middleware extracts schema
- `/health` -> Health check endpoint
- `/swagger/*` -> Swagger UI (`shared/swagger/chi.go`)
- `/*` -> Reverse proxy to Next.js on :3001

### Inter-Module Communication

- **Synchronous:** gRPC (only `organization` module has full implementation; proto in `infrastructure/grpc/proto/`)
- **Asynchronous:** EventBus (in-memory, `shared/eventbus/`). Events published from `infrastructure/messaging/`. Currently only `organization` module uses this.
- **SSE:** Pub/sub brokers (`shared/pubsub/`) for real-time notifications — `CheckBroker` (page check status) and `InsightBroker` (insight generation completion)
- Never share structs between modules; deserialize to own types

### Shared Packages (`shared/`)

Technical utilities only — no business logic. 15 packages:

| Package | Description |
|---------|-------------|
| `ai/` | OpenRouter LLM client — text completions and multimodal (vision) analysis |
| `bff/` | BFF auth routes — login, refresh, logout, callback, set-base-session (cookie + nonce management) |
| `cache/` | Redis client singleton + refresh token cache (2s grace period for concurrent refresh) |
| `config/` | Environment variable loading with defaults (43+ config vars, see `.env.example`) |
| `database/` | PostgreSQL connection pool (exponential backoff retry) + tenant schema provisioning/migration |
| `eventbus/` | In-memory pub/sub (`MessageBus` interface — swappable for Kafka in production) |
| `html/` | HTML text extraction — DOM tree walker, strips scripts/styles, normalizes whitespace |
| `http/` | Response helpers (`RespondJSON`, `RespondError`, status shortcuts) + Chi middleware (logging, recovery, requestID, timeout, CORS) |
| `logger/` | Zap structured logging — context-aware functions with correlation_id, tenant, user_id extraction |
| `middleware/` | HTTP middleware: tenant extraction, auth provider, organization membership, rate limiter (token bucket per IP), request logging |
| `noncestore/` | In-memory nonce store (30s TTL) for cross-subdomain token exchange |
| `pubsub/` | SSE brokers — `CheckBroker` (2min cache TTL) and `InsightBroker` (5min cache TTL, one-shot delivery) |
| `router/` | Module registration registry (`ModuleRegisterer` interface) |
| `static/` | Reverse proxy to Next.js (dev) or static file serving (prod), with subdomain -> X-Tenant header injection |
| `swagger/` | Swagger UI setup for Chi — serves doc.json and Swagger UI assets |

### Database Migrations

- Public schema: `shared/database/migrations/public/` — 12 migrations (users, orgs, sessions, roles, permissions, plans, OAuth providers, invitation status)
- Tenant schema: `shared/database/migrations/tenant/` — 12 migrations (workspaces, pages, checks, insights, monitoring configs, usage tracking, monitored sections, section rects, parent check relationships)
- Format: `000001_description.up.sql` / `000001_description.down.sql`
- Tenant migrations apply to all tenant schemas uniformly via `golang-migrate/migrate/v4`
- Scaffold new migrations: `./tools/scripts/new-migration.sh <scope> <description>`

### Background Jobs

Worker process (`cmd/worker/` or `ENABLE_WORKERS=true` in monolith) handles:
- Scheduled monitoring checks (30s poll interval via `infrastructure/scheduler/`)
- Concurrent check execution via worker pool (`application/workers/`)
- Orchestration of snapshot -> change detection -> insight generation -> alerts (`application/orchestrator/`)
- Snapshot capture via Playwright scraper service (HTTP client in `snapshot/infrastructure/extractor/`)
- Object storage upload to MinIO or Cloudinary (`snapshot/infrastructure/minio/`, `snapshot/infrastructure/cloudinary/`)
- AI insight generation via OpenRouter (`insight/infrastructure/ai/`)
- Alert creation and email notifications via Resend
- Webhook publishing to configured integrations

### Frontend (`frontend/`)

Bun workspace monorepo with Turborepo:

#### Apps
- `apps/web/` — Next.js 16 App Router application (port :3001)

#### Packages
| Package | Name | Description |
|---------|------|-------------|
| `packages/ui/` | `@workspace/ui` | Atomic Design component library (atoms/molecules/organisms) using Radix UI + Tailwind |
| `packages/services/` | `@workspace/services` | API client layer — 12 service files (auth, workspace, page, dashboard, notification, organization, team, usage, super-admin, integration, report) |
| `packages/shared-http/` | `@workspace/shared-http` | Tenant-aware HTTP factory (Axios for browser, Fetch for SSR), `IHttpClient` interface, subdomain extraction |
| `packages/notix/` | `@workspace/notix` | Toast notification library (hexagonal architecture, motion animations) |
| `packages/typescript-config/` | `@workspace/typescript-config` | Shared TSConfig presets (base, nextjs, react-library) |

#### Route Groups (16 pages, 4 layouts)
- `(auth)/` — Login, invite acceptance (redirects if already authenticated)
- `(main)/` — Authenticated app wrapped in `AuthGuard` + `AppShell` (sidebar + header): dashboard, workspaces, workspace detail, page detail, changes view, reports, settings, team, admin
- `(public)/` — Registration (no auth required)
- `(demo)/` — Experimental pages (lecture-ai)

#### Features (17 vertical slices)
`features/{name}/` with UI, application, domain layers:
account-settings, auth, changes-view, dashboard, landing, navigation, notifications, page, page-detail, reports, settings, sidebar, super-admin, team, usage, workspace, workspace-detail

#### Auth Protection
- No Next.js middleware file — auth is handled by `AuthGuard` (async React Server Component that calls `AuthApi.getCurrentUser()` and redirects to `/login` on unauthorized)
- `(auth)` layout checks if user is already authenticated and redirects to tenant subdomain

### Docker & Deployment

#### Docker Compose (`docker-compose.monolith.yml`)
5 services: `postgres` (17-alpine, port 5434), `localstack` (S3 emulator, port 4566), `extractor` (Playwright, port 3005), `monolith` (API with air hot reload, port 3000), `worker` (background worker with air hot reload)

#### Dockerfiles
| File | Purpose |
|------|---------|
| `Dockerfile.api` | Production API server (multi-stage, non-root, healthcheck) |
| `Dockerfile.worker` | Production worker (multi-stage, non-root) |
| `Dockerfile.monolith.all-in-one` | Development monolith (air hot reload + swag) |
| `modules/infra/scraper/Dockerfile` | Playwright scraper/extractor (Bun + Chromium) |
| `modules/snapshot/Dockerfile` | Snapshot worker with Chromium |
| `frontend/Dockerfile` | Production Next.js (multi-stage Turborepo build) |

#### Railway (production)
4 service configs in `railway/`: api, worker, extractor, frontend — each with Dockerfile builder and restart policy (ON_FAILURE, 3 retries).

### Key Config Variables

See `shared/config/config.go` and `.env.example` for all 43+ variables. Critical groups:

| Category | Variables | Notes |
|----------|-----------|-------|
| **Database** | `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` | Required. Also supports `DATABASE_URL` (Railway) |
| **Server** | `HTTP_PORT` (default 3000), `GRPC_PORT` (default 9000), `ENVIRONMENT`, `ENABLE_WORKERS` | |
| **Auth (JWT)** | `JWT_SECRET`, `JWT_EXPIRATION` (15min), `JWT_REFRESH_EXPIRATION` (7d) | `JWT_SECRET` required in production |
| **CORS** | `CORS_ALLOWED_ORIGINS` | Required. Comma-separated |
| **Cookie** | `COOKIE_DOMAIN` | Cross-subdomain auth scope |
| **Frontend** | `FRONTEND_URL`, `NEXTJS_URL` (default localhost:3001), `STATIC_DIR` | |
| **Extractor** | `EXTRACTOR_URL` | Required. Playwright service URL |
| **Object Storage** | `OBJECT_STORAGE_PROVIDER` (minio/cloudinary), `MINIO_*` (6 vars), `CLOUDINARY_*` (4 vars) | |
| **AI Insights** | `OPENROUTER_API_KEY`, `OPENROUTER_MODEL`, `OPENROUTER_VISION_MODEL`, `PIXEL_DIFF_THRESHOLD` | Optional |
| **Email** | `RESEND_API_KEY`, `EMAIL_FROM_ADDRESS`, `EMAIL_FROM_NAME` | Optional |
| **OAuth** | `GOOGLE_CLIENT_ID/SECRET`, `GITHUB_CLIENT_ID/SECRET`, `OAUTH_REDIRECT_BASE_URL` | Optional |
| **Redis** | `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD` | Optional (graceful degradation) |
| **Rate Limiting** | `RATE_LIMIT_REQUESTS` (500), `RATE_LIMIT_WINDOW` (60s) | |
| **Logging** | `LOG_LEVEL` (debug/info/warn/error) | |

### Architecture Improvements & Scaling Strategy

#### Cross-Module Dependency Violations

Several modules import directly from other modules' infrastructure layers, violating the hexagonal boundary:
- `monitoring` imports from `insight`, `snapshot`, `email` infrastructure
- `insight` imports from `monitoring` infrastructure (persistence repos)
- `admin` imports from `auth`, `email`, `organization` infrastructure
- `auth` imports from `admin`, `email`, `organization` infrastructure
- `team` imports from `auth`, `email` infrastructure

**Recommended fix:** Define shared interfaces in `shared/` or use an anti-corruption layer. Each module should depend only on interfaces, not concrete implementations from other modules. Wire implementations via dependency injection in `cmd/server/modules.go`.

#### Modules Needing Refactoring

- **`usage`** — All logic (~713 lines) is inline in `module.go` with empty domain/application directories. Extract into proper hexagonal layers: entities (`Plan`, `UsageTracking`, `BillingPeriod`), repository interfaces, and use cases (`get_metrics`, `get_quotas`, `assign_plan`, `gift_month`).
- **`report`** — All handlers inline in `module.go`, empty `create_report/` directory. Extract into use cases with proper request/response DTOs.
- **`auth`** — Several inline handlers (`forgot_password`, `reset_password`, `update_current_user`, `change_password`, `delete_current_user`) should be extracted into dedicated use case directories.

#### Performance & Scaling

- **Database connection pooling:** Currently uses a single `*sql.DB` pool shared across all tenants. For high tenant counts, consider per-tenant connection pooling or a connection pool proxy (PgBouncer).
- **`SET search_path` per query:** Every tenant-scoped query prepends `SET search_path`. Consider using `schema_name.table_name` qualified queries to eliminate this overhead, or use connection-level schema setting.
- **EventBus:** The in-memory `EventBus` is not durable and loses events on restart. For production reliability, implement the Kafka adapter behind the `MessageBus` interface. The interface is already designed for this swap.
- **Redis:** Currently optional with graceful degradation. For production, Redis should be required for refresh token deduplication, rate limiting state sharing across instances, and session caching.
- **Horizontal scaling:** The monolith can scale horizontally because tenant state is in PostgreSQL. However, the in-memory `NonceStore`, `EventBus`, `CheckBroker`, and `InsightBroker` are node-local. Replace with Redis-backed implementations for multi-instance deployments.
- **Worker scaling:** The worker process (`cmd/worker/`) is a single instance. For higher throughput, implement distributed job locking (Redis/PostgreSQL advisory locks) to allow multiple worker instances.
- **Scraper concurrency:** The scraper limits to `MAX_CONCURRENT_PAGES` (default 3). Scale by running multiple scraper instances behind a load balancer, with the Go backend round-robining requests.

#### Security Improvements

- **Scraper SSRF:** The scraper has no URL validation beyond checking presence. Add URL protocol validation (only `http://` and `https://`), block private/internal IPs, and consider adding an API key for service-to-service auth.
- **Scraper request limits:** No body size limits on `/extract` and `/preview` endpoints. Add max body size middleware.

#### Legacy Cleanup

- 11 modules still have legacy standalone artifacts (`main.go`, `.air.toml`, `docs/`, `tmp/`) from the pre-monolith era. These should be removed.
- `organization` module has a stale compiled binary and legacy Gin `router.go` — remove.
- `shared/middleware/health.go` uses Gin (not Chi) — likely unused, should be removed or ported.
- Empty placeholder directories (`integration/setup_webhook/`, `snapshot/kafka/`, `report/create_report/`, `usage/track_usage/`, `usage/entities/`, `usage/repositories/`) should either be implemented or removed.
