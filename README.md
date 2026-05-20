# Pulzifi

Website monitoring platform with AI-powered change detection and insights. Monitors web pages on configurable schedules, detects visual and content changes via Playwright screenshots, and generates intelligent summaries using LLM analysis.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend API | Go 1.25, Chi router |
| Frontend | Next.js 16, React 19, Tailwind CSS |
| Database | PostgreSQL 17 (schema-per-tenant) |
| Screenshots | Playwright (Node.js extractor service) |
| AI Insights | OpenRouter (OpenAI-compatible API) |
| Email | Resend |
| Storage | S3-compatible (MinIO / LocalStack dev) |
| Package Manager | Bun 1.1 + Turborepo |
| Linter | Biome |

## Repository Structure

```
cmd/
  server/          HTTP :3000 + gRPC :50051 entry point
  worker/          Background job runner
  migrate/         Database migration CLI
modules/           Business domain modules (hexagonal architecture)
  auth/            Authentication, JWT, OAuth2
  organization/    Org management, subdomain provisioning
  workspace/       Workspace CRUD + member roles
  page/            Page URL management
  monitoring/      Check scheduling, SSE streaming
  snapshot/        Playwright capture, change detection
  insight/         AI-powered analysis
  alert/           Change notifications
  team/            Member invitations
  email/           Email templates + Resend provider
  integration/     Webhooks, Slack, Discord
  dashboard/       Aggregated statistics
  admin/           User approval workflow
  report/          Report generation
  usage/           Usage tracking
  infra/extractor/ Playwright Node.js service
shared/            Cross-cutting utilities (no business logic)
  config/          Environment variable loading
  database/        PostgreSQL connection + migrations
  middleware/      Tenant extraction, JWT, rate limiting
  bff/             Backend-for-Frontend auth orchestration
  noncestore/      Cross-subdomain token exchange
  pubsub/          SSE brokers for real-time updates
  ai/              OpenRouter LLM client
  eventbus/        In-memory pub/sub
  logger/          Zap structured logging
  html/            HTML text extraction
  cache/           Redis client
  router/          Module registration
  static/          Frontend proxy / static file serving
frontend/
  apps/web/        Next.js application (App Router)
  packages/ui/     Atomic Design component library
  packages/services/   API client layer
  packages/shared-http/ Tenant-aware HTTP factory
```

## Prerequisites

- Go 1.25+
- Docker & Docker Compose
- Bun 1.1+
- Node.js 22+ (for extractor service)

## Quick Start

1. **Clone and configure**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials and JWT secret
   ```

2. **Install frontend dependencies** (required for `cms-migrate` and `dev-web`)
   ```bash
   cd frontend && bun install && cd ..
   ```

3. **Start PostgreSQL first** (required before migrations)
   ```bash
   docker-compose up -d postgres
   # Wait ~5s until healthy
   ```

4. **Run migrations** (public + tenant schemas)
   ```bash
   make migrate
   make cms-migrate   # only if using Payload CMS — requires step 2
   ```

   Skipping migrations makes the monolith crash on startup — schemas/tables won't exist yet.

5. **Start the full development stack**
   ```bash
   make dev      # Starts scraper, API server, and worker (postgres already up)
   ```

6. **Start the frontend** (separate terminal)
   ```bash
   make dev-web  # Starts Next.js on :3001 (Go proxies from :3000)
   ```

7. **Access the application**
   - App: `http://localhost:3000`
   - Swagger: `http://localhost:3000/swagger/`

### Subsequent runs

Once volumes exist, just:
```bash
make dev                    # full stack
make migrate                # only if new migrations were pulled
```

### Full reset

```bash
make down                   # drops volumes
docker-compose up -d postgres
make migrate
make dev
```

## Development Commands

| Command | Description |
|---------|-------------|
| `make dev` | Start full dev stack (postgres + extractor + API + worker) |
| `make dev-web` | Start Next.js on :3001 |
| `make down` | Stop dev environment and remove volumes |
| `make build` | Build Go binary to `./bin/api` |
| `make swagger` | Regenerate Swagger docs |
| `make migrate` | Run all migrations |
| `make migrate cmd=down` | Rollback migrations |
| `make clean` | Stop containers and prune Docker |

### Running Tests

```bash
go test ./...                        # All tests
go test ./modules/workspace/...      # Single module
go test -v -run TestName ./path/...  # Single test
go test -race ./...                  # With race detector
```

### Frontend

```bash
cd frontend
bun dev            # Dev server on :3001
bun run build      # Production build
bun run lint:fix   # Format + lint (Biome)
```

## Git Hooks (commit + push gates)

The repo ships tracked hooks under `.githooks/`. Run once per clone:

```bash
bash tools/scripts/install-hooks.sh    # sets core.hooksPath -> .githooks
```

| Hook | Script | Checks |
|------|--------|--------|
| `pre-commit` | `tools/scripts/validate-build.sh` | `go build`, `go vet`, frontend `type-check` |
| `pre-push`   | `tools/scripts/pre-push-validate.sh` | `go test`, `check-architecture.sh`, `bun test features/cms`, frontend `type-check`, frontend `lint` |

The pre-push gate is the global error-catching net — fast feedback before the push leaves the machine. Failed checks print as `[name] <output>`, capped at 30 lines.

## CMS (Payload) — block-ref invariant

The Payload CMS seed wires `pages.blocks` to entries in the `block-library` collection by name. If a referenced block disappears, Payload rejects every page save with `"The following field is invalid: id"`.

Guard rails in place:

- **Fail-fast seed** — `frontend/apps/web/features/cms/seed.ts` `ref()` throws when a `block-library` name is missing instead of writing `ref: undefined`.
- **Rebuild by default** — `seedCMSIfEmpty` always re-runs `seedAll`, keeping page block-refs aligned with current `block-library` ids on every deploy. Set `CMS_PRESERVE_EDITS=true` once real editor work begins to protect manual changes.
- **`validateRefs(payload)`** — walks every page, asserts each `block-ref.ref` resolves in `block-library`. Used by the bun test suite and available for ad-hoc debugging.
- **Pre-push test gate** — `bun test features/cms` runs the seed + validator against an in-memory fake Payload on every push.

## Architecture Overview

- **Hexagonal Architecture + Vertical Slicing**: Each module is self-contained with domain, application, and infrastructure layers
- **Multi-Tenant by Subdomain**: `tenant.app.com` resolves to a dedicated PostgreSQL schema
- **Single HTTP Entry Point**: Go on :3000 serves API routes, BFF auth, and proxies to Next.js
- **Inter-Module Communication**: gRPC (sync) + EventBus/Kafka (async)
- **Real-Time**: SSE streams for check status and insight generation

See [docs/architecture.md](docs/architecture.md) for detailed system design.

## Environment Variables

See [.env.example](.env.example) for all options. Required:

| Variable | Description |
|----------|-------------|
| `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` | PostgreSQL connection |
| `JWT_SECRET` | JWT signing key |
| `CORS_ALLOWED_ORIGINS` | Allowed origins (comma-separated) |
| `EXTRACTOR_URL` | Playwright screenshot service URL |
| `OBJECT_STORAGE_PROVIDER` | `minio` or `cloudinary` |

Optional:

| Variable | Description |
|----------|-------------|
| `OPENROUTER_API_KEY` | AI insights (disabled if unset) |
| `RESEND_API_KEY` | Email sending (disabled if unset) |
| `GOOGLE_CLIENT_ID` / `GITHUB_CLIENT_ID` | OAuth providers |

## License

Proprietary. All rights reserved.
