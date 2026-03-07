# Entry Points (cmd/)

Three CLI entry points for the Pulzifi backend.

## cmd/server/

Main HTTP + gRPC monolith server.

### Files
- `main.go` — Application bootstrap: config loading, database/Redis connection, EventBus init, Chi router setup, CORS, rate limiting, route mounting, gRPC server, graceful shutdown
- `modules.go` — Module registration and dependency wiring: instantiates all 14 modules, creates repositories and services, configures auth middleware, builds BFF handler

### Responsibilities
- Starts HTTP server on `:3000` (configurable via `HTTP_PORT`)
- Starts gRPC server on `:9000` (configurable via `GRPC_PORT`)
- Registers 14 module routes under `/api/v1/*`
- Mounts BFF auth at `/api/auth/*`
- Proxies unmatched routes to Next.js on `:3001`
- If `ENABLE_WORKERS=true`, also runs background monitoring processes (all-in-one mode)
- Handles graceful shutdown on SIGINT/SIGTERM

### Module Registration Order
Auth, Admin, Email, Organization, Workspace, Page, Alert, Monitoring, Integration, Insight, Report, Usage, Dashboard, Team

## cmd/worker/

Standalone background worker process (49 lines).

### Responsibilities
- Loads config and connects to database
- Instantiates the Monitoring module (with nil EventBus and email provider)
- Calls `StartBackgroundProcesses()` on the monitoring module
- Blocks on SIGINT/SIGTERM for graceful shutdown

### Notes
- Minimal entry point — all logic lives in the monitoring module's orchestrator and worker pool
- Used when deploying API and worker as separate services (production)
- In development, `ENABLE_WORKERS=true` on the server makes this unnecessary

## cmd/migrate/

Database migration CLI tool.

### Flags
- `-db` — Database URL (defaults from environment variables via godotenv)
- `-cmd` — Migration command: `up`, `down`, `version`, `force`
- `-steps` — Number of migration steps (for partial up/down)
- `-scope` — Migration scope: `all` (default), `public`, `tenant`
- `-tenant` — Specific tenant schema name

### Behavior
- Validates schema names with regex `^[a-zA-Z_][a-zA-Z0-9_]*$`
- For scope `all`: runs public migrations first, then all tenant schemas
- For scope `tenant` with `-tenant` flag: runs only that specific schema
- Without `-tenant`: queries all tenant schemas from `organizations` table
- Uses `golang-migrate/migrate/v4` with file-based migration sources

### Examples
```bash
go run ./cmd/migrate -cmd up                          # All migrations
go run ./cmd/migrate -scope public -cmd up            # Public schema only
go run ./cmd/migrate -scope tenant -tenant demo -cmd up  # Specific tenant
go run ./cmd/migrate -cmd down -steps 1               # Rollback 1 step
go run ./cmd/migrate -cmd version                     # Check current version
```
