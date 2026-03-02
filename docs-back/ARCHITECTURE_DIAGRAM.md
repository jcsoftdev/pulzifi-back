# Pulzifi Backend - Architecture Diagrams

## System Overview

```
┌──────────────────────────────────────────────────────────────┐
│                    Go HTTP Server (:3000)                     │
│                   (Single Entry Point)                        │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │              Chi Router (Port 3000)                    │  │
│  │  CORS ─► Rate Limiter ─► Route Matching               │  │
│  └───────────────────┬────────────────────────────────────┘  │
│                      │                                       │
│     ┌────────────────┼──────────────────┐                    │
│     │                │                  │                    │
│     ▼                ▼                  ▼                    │
│  /api/auth/*     /api/v1/*            /*                     │
│  BFF Handler     Module Routes        Reverse Proxy          │
│  (cookies,       (tenant MW ──►       to Next.js :3001       │
│   nonces)         module handlers)                           │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │              gRPC Server (Port 9000)                   │  │
│  │        (Inter-module communication)                    │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │     Shared Services                                    │  │
│  │  - Database Pool (PostgreSQL, schema-per-tenant)       │  │
│  │  - Redis Cache (optional)                              │  │
│  │  - EventBus (in-memory pub/sub)                        │  │
│  │  - Logger (Zap structured logging)                     │  │
│  │  - Rate Limiter                                        │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

---

## Module Registry

All 17 modules register via the `ModuleRegisterer` interface:

```
cmd/server/main.go
        │
        ├─ config.Load()
        ├─ database.Connect()
        ├─ cache.InitRedis()
        ├─ eventbus.GetInstance()
        │
        ├─ registerAllModulesInternal(registry, db, eventBus, enableWorkers)
        │       │
        │       ├─ Auth          (users, JWT, OAuth)
        │       ├─ Admin         (registration approval)
        │       ├─ Email         (Resend templates)
        │       ├─ Organization  (orgs, subdomains)
        │       ├─ Workspace     (workspace CRUD, members)
        │       ├─ Page          (URL registration)
        │       ├─ Alert         (change alerts)
        │       ├─ Monitoring    (check scheduling, workers)
        │       ├─ Integration   (webhooks)
        │       ├─ Insight       (LLM analysis)
        │       ├─ Report        (reporting)
        │       ├─ Usage         (plan limits)
        │       ├─ Dashboard     (org statistics)
        │       ├─ Team          (member invitations)
        │       └─ Snapshot      (extractor service)
        │
        ├─ Mount BFF routes at /api/auth
        ├─ Mount module routes at /api/v1 (with tenant middleware)
        ├─ Mount reverse proxy for /* → Next.js
        │
        ├─ Start HTTP server (:3000)
        └─ Start gRPC server (:9000)
```

---

## Request Flow

### API Request (Module Routes)

```
Client Request
    │
    ▼
┌─────────────────────┐
│ Go Server (:3000)   │
│ Chi Router          │
└────────────┬────────┘
             │
             ▼
    ┌────────────────┐
    │ CORS Middleware │
    └────────┬───────┘
             │
    ┌────────▼────────┐
    │ Rate Limiter    │
    └────────┬────────┘
             │
    ┌────────▼──────────────┐
    │ Route: /api/v1/*      │
    └────────┬──────────────┘
             │
    ┌────────▼────────────────┐
    │ Tenant Middleware       │
    │ subdomain → schema_name │
    └────────┬────────────────┘
             │
    ┌────────▼──────────────┐
    │ Response Logger       │
    └────────┬──────────────┘
             │
    ┌────────▼──────────────┐
    │ Module Handler        │
    │ (Auth, Page, etc.)    │
    └────────┬──────────────┘
             │
             ▼
      Response to Client
```

### BFF Authentication Request

```
Client Request
    │
    ▼
┌─────────────────────────────┐
│ Route: /api/auth/*          │
└────────────┬────────────────┘
             │
    ┌────────▼────────────────┐
    │ BFF Handler             │
    │ (shared/bff/handler.go) │
    └────────┬────────────────┘
             │
    ┌────────┼────────────────────┐
    │        │                    │
    ▼        ▼                    ▼
  /login   /callback           /refresh
    │        │                    │
    ▼        ▼                    ▼
  Auth     Consume nonce       Validate
  Module → Store in           refresh
  Login    NonceStore          cookie →
  Handler  (30s TTL) →        New tokens →
    │      Set HttpOnly        Set cookies
    ▼      cookies on
  Return   tenant subdomain
  nonce
```

### Frontend Request (Reverse Proxy)

```
Client Request (not /api/*)
    │
    ▼
┌───────────────────────────┐
│ Go Server (:3000)         │
│ No matching /api/* route  │
└────────────┬──────────────┘
             │
    ┌────────▼──────────────────────┐
    │ Reverse Proxy                 │
    │ (shared/static/handler.go)    │
    │ FlushInterval: -1 (SSE/HMR)  │
    └────────────┬──────────────────┘
                 │
    ┌────────────▼──────────────┐
    │ Next.js Dev Server (:3001)│
    │ (pages, components, SSR)  │
    └───────────────────────────┘
```

---

## Data Flow

### Monitoring Check Lifecycle

```
Page created → MonitoringConfig created (frequency, timezone)
                        │
              Scheduler polls every 30s
                        │
              Finds due configs → dispatches to WorkerPool
                        │
              Worker calls Extractor (Playwright :3005)
                        │
              Receives screenshot + HTML
                        │
              Computes contentHash (SHA256 of HTML)
                        │
              Compares with previous check
                        │
              ┌─── no change ──→ Store check (changeDetected=false)
              │
              └─── change detected ──→ Store check (changeDetected=true)
                                              │
                                    ┌─────────┼─────────┐
                                    │         │         │
                              Create Alert  SSE Push  Generate Insights
                                              │         │
                                              │   LLM analysis (OpenRouter)
                                              │         │
                                              │   Store insights
                                              │         │
                                              └── SSE notify client
```

### Authentication Flow (Cross-Subdomain)

```
1. User visits app.domain.com/login
2. POST /api/auth/login → BFF handler
3. BFF calls auth module login handler
4. On success: generate nonce, store tokens in NonceStore (30s TTL)
5. Return nonce + tenant to frontend
6. Frontend redirects to tenant.domain.com/api/auth/callback?nonce=<uuid>
7. BFF consumes nonce → sets HttpOnly cookies on tenant subdomain
8. Redirect to tenant.domain.com/workspaces
```

---

## Module Interface

```go
// shared/router/registry.go
type ModuleRegisterer interface {
    RegisterHTTPRoutes(router chi.Router)
    ModuleName() string
}
```

Each module implements this in `modules/<name>/infrastructure/http/module.go`:

```go
type Module struct {
    // Dependencies injected via constructor
}

func NewModule(deps ModuleDeps) router.ModuleRegisterer {
    return &Module{...}
}

func (m *Module) ModuleName() string {
    return "ModuleName"
}

func (m *Module) RegisterHTTPRoutes(r chi.Router) {
    r.Route("/resource", func(r chi.Router) {
        r.Use(middleware.AuthMiddleware())
        r.Get("/", m.handleList)
        r.Post("/", m.handleCreate)
        // ...
    })
}
```

---

## Dependency Graph

```
┌──────────────────────────┐
│    Shared Libraries      │
├──────────────────────────┤
│ - config                 │
│ - database               │
│ - logger                 │
│ - middleware             │
│ - router                 │
│ - bff                    │
│ - noncestore             │
│ - pubsub                 │
│ - eventbus               │
│ - cache                  │
│ - ai                     │
│ - html                   │
│ - static                 │
│ - swagger                │
│ - http (shared helpers)  │
└──────────────────────────┘

┌──────────────────────────┐
│   Module (each one)      │
├──────────────────────────┤
│ - domain                 │
│   ├─ entities            │
│   ├─ repositories (iface)│
│   ├─ services            │
│   ├─ errors              │
│   └─ value_objects       │
│ - application            │
│   └─ <use_case>/         │
│       ├─ handler.go      │
│       ├─ request.go      │
│       ├─ response.go     │
│       └─ handler_test.go │
│ - infrastructure         │
│   ├─ http (module.go)    │
│   ├─ persistence         │
│   ├─ grpc                │
│   └─ messaging           │
└──────────────────────────┘

Orchestrator:
┌──────────────────────────┐
│  cmd/server/main.go      │
│  cmd/server/modules.go   │
│  (Wires all modules)     │
└──────────────────────────┘
```

---

## File Structure

```
pulzifi-back/
├── cmd/
│   ├── server/
│   │   ├── main.go              # HTTP + gRPC entry point
│   │   └── modules.go           # Module registration and wiring
│   ├── worker/
│   │   └── main.go              # Background job entry point
│   └── migrate/
│       └── main.go              # Database migration runner
│
├── modules/                     # 17 domain modules
│   ├── admin/                   # Registration approval
│   ├── alert/                   # Change alerts
│   ├── auth/                    # Authentication, JWT, OAuth
│   ├── dashboard/               # Org-wide statistics
│   ├── email/                   # Email templates (Resend)
│   ├── infra/                   # Playwright extractor (Node.js)
│   ├── insight/                 # LLM-powered analysis
│   ├── integration/             # Third-party webhooks
│   ├── monitoring/              # Check scheduling, workers
│   ├── organization/            # Org lifecycle, subdomains
│   ├── page/                    # URL registration
│   ├── report/                  # Report generation
│   ├── snapshot/                # Screenshot capture, change detection
│   ├── team/                    # Member invitations
│   ├── usage/                   # Resource usage tracking
│   └── workspace/               # Workspace CRUD, member roles
│
├── shared/                      # Technical utilities (no business logic)
│   ├── ai/                      # OpenRouter LLM client
│   ├── bff/                     # BFF auth handler
│   ├── cache/                   # Redis client
│   ├── config/                  # Env var loading
│   ├── database/                # PostgreSQL pool + migrations
│   ├── eventbus/                # In-memory pub/sub
│   ├── html/                    # HTML text extraction
│   ├── http/                    # Shared HTTP helpers
│   ├── logger/                  # Zap structured logging
│   ├── middleware/              # Tenant, JWT, rate limiter, org
│   ├── noncestore/             # In-memory nonce store (30s TTL)
│   ├── pubsub/                  # SSE brokers (Check, Insight)
│   ├── router/                  # Module registry
│   ├── static/                  # Reverse proxy to frontend
│   └── swagger/                 # Swagger UI setup
│
├── frontend/                    # Next.js + Turborepo + Bun
│   ├── apps/web/                # App Router application
│   └── packages/                # Shared UI, services, HTTP
│
├── docs-back/                   # Backend architecture docs
├── docs/                        # ADRs, runbooks
├── docker-compose.monolith.yml  # Dev stack (Postgres, LocalStack)
└── Makefile                     # Dev, build, migrate commands
```
