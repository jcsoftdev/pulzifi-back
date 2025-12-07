# .github/copilot-instructions.md

## Copilot Preferences

- No summaries or documentation files
- No comprehensive guides of the implementations
- Focus only on code implementation
- Direct fixes without explanations
- Minimal console output
- Run commands if you need it

## Project Architecture

**Stack:** Go + Hexagonal + Vertical Slicing + Screaming Architecture + Multi-Tenant by subdomain

**Principle:** Each module is independent and deployable. Modules communicate as external services.

---

## Structure

```
pulzifi-back/
  shared/                               # Cross-cutting utilities (NO business logic)
    config/
      loader.go                         # Load environment variables
    database/
      connection.go                     # PostgreSQL connection pool
      migrations/
        public/                         # Public schema migrations
          001_create_users.up.sql
          002_create_organizations.up.sql
    middleware/
      tenant_extractor.go               # Extract tenant from X-Tenant header
      jwt_validator.go                  # Validate JWT and extract user_id
      logger.go                         # Request logging
    logger/
      logger.go                         # Structured logging (zerolog/zap)

  modules/
    
    # ============================================================
    # EXAMPLE: workspace module (tenant schema)
    # ============================================================
    workspace/
      
      # HEXAGONAL - CORE (Domain + Application)
      domain/
        entities/
          workspace.go                  # Workspace entity
        value_objects/
          workspace_type.go             # "Personal", "Competitor"
        repositories/
          workspace_repository.go       # Repository interface
        events/
          workspace_created.go          # Domain event
          workspace_deleted.go
        services/
          workspace_service.go          # Shared domain logic
        errors/
          workspace_errors.go           # Business errors
      
      application/                      # Use cases (VERTICAL SLICING)
        create_workspace/
          handler.go                    # Orchestrates: validates, creates, publishes event
          request.go                    # Input DTO
          response.go                   # Output DTO
          handler_test.go
        list_workspaces/
          handler.go
          request.go                    # Filters, pagination
          response.go
          handler_test.go
        get_workspace/
          handler.go
          response.go
          handler_test.go
        update_workspace/
          handler.go
          request.go
          response.go
          handler_test.go
        delete_workspace/
          handler.go
          response.go
          handler_test.go
        get_workspace_statistics/
          handler.go
          response.go                   # Stats: page count, daily checks
          handler_test.go
      
      # HEXAGONAL - ADAPTERS (Infrastructure)
      infrastructure/
        http/                           # REST API for frontend
          router.go                     # HTTP routes: POST /workspaces, GET /workspaces/:id
          middleware.go                 # Extract tenant, validate JWT
          handlers/
            create_workspace_handler.go # Adapts HTTP → application/create_workspace
            list_workspaces_handler.go
            get_workspace_handler.go
            update_workspace_handler.go
            delete_workspace_handler.go
        
        grpc/                           # gRPC for inter-module communication
          proto/
            workspace.proto             # gRPC definition
          server.go                     # gRPC server implementation
          interceptors.go               # Extract tenant from metadata
          organization_client.go        # Client to call organization module
        
        persistence/
          workspace_postgres.go         # Repository implementation (PostgreSQL)
          workspace_memory.go           # In-memory implementation (for tests)
          mapper.go                     # Mapping between DB models and domain entities
        
        messaging/
          publisher.go                  # Publish events to Kafka
          subscriber.go                 # Subscribe to events (if applicable)
      
      main.go                           # Bootstrap: HTTP server (8082) + gRPC server (9082)

```

---

## Dependency Rules

1. `domain` imports nothing
2. `application` only imports `domain`
3. `infrastructure` imports `application` and `domain`
4. Interfaces in `domain/repositories/`, implementations in `infrastructure/persistence/`
5. **No imports between modules**
6. Communication: gRPC (synchronous) or Kafka (asynchronous)

---

## Multi-Tenant

**Extraction:**
- Always from subdomain: `tenant1.app.com` (production and development)

**Usage:**
```go
ctx = context.WithValue(ctx, TenantKey, tenant)
tenant := ctx.Value(TenantKey).(string)
db.Exec("SET search_path TO " + tenant)
```

**Rules:**
- Validate tenant in middleware/interceptor
- Pass tenant explicitly in all operations
- Never use "public" or hardcode tenant
- If tenant doesn't exist: return 404

---

## Inter-Module Communication

### gRPC (synchronous)
- Proto in `infrastructure/grpc/proto/<module>.proto`
- Server in `infrastructure/grpc/server.go`
- Client in `infrastructure/grpc/<module>_client.go`
- Interceptor to inject tenant in metadata

### Kafka (asynchronous)
- Publisher in `infrastructure/messaging/publisher.go`
- Subscriber in `infrastructure/messaging/subscriber.go`
- Events in `domain/events/` (types only)
- Format: JSON
- **Do not share structs**, deserialize to own types
- Include tenant in message

---

## Naming

- Features: `create_invoice/`, `get_user/`, `send_notification/`
- DTOs inside feature
- Entities singular: `User`, `Invoice`
- Shared module logic → `domain/services/`
- **Do not create** `application/shared/`

---

## Shared/ - Technical Only

**✅ YES:**
- config, database, middleware, logger
- Technical errors (DatabaseError, ConfigError)

**❌ NO:**
- Business logic, repositories, DTOs, mocks, domain errors

---

## main.go - Bootstrap Only

**✅ YES:**
- Load config, connect DB
- Instantiate repos, handlers
- Start HTTP server (REST API)
- Start gRPC server (inter-module communication)
- Start Kafka consumer

**❌ NO:**
- Validations, queries, business logic

---

## Testing

- Tests in `*_test.go` alongside code
- Unit: use `<entity>_memory.go`
- Integration: testcontainers
- gRPC: use mock client

---

## Transactions

- Start in `application/handler.go`
- Pass `*sql.Tx` to repo
- Commit/Rollback in handler

---

## Migrations

### Public Schema (Shared)
- Location: `shared/database/migrations/public/`
- Contains: users, organizations, organization_members, refresh_tokens, password_resets
- Includes function: `create_tenant_schema()` with complete tenant structure

### Tenant Schema (Per Organization)
- NO migrations per module (all tenants use same structure)
- The `create_tenant_schema()` function runs automatically when creating an organization
- Contains: workspaces, pages, checks, alerts, insights, reports, etc.

### Format
- `001_create_table.up.sql` / `001_create_table.down.sql`

---

## Background Jobs

### Asynq (Redis-based) - For scheduled tasks
- Scheduled monitoring checks
- Email sending with retry logic
- AI insight generation
- Usage quota refill (monthly)

### Kafka Consumers - For inter-module events
- `check_completed` → alert, insight, usage modules
- `alert_created` → integration module
- Each module subscribes only to relevant events

---

## Network Infrastructure

### Load Balancer / API Gateway (Nginx/Traefik/Kong)
**NOT a code module, it's infrastructure:**
- Terminates SSL/TLS
- Extracts subdomain and passes it as header (`X-Tenant`)
- Routes requests to modules by path:
  - `/api/auth/*` → auth module
  - `/api/workspaces/*` → workspace module
  - `/api/pages/*` → page module
  - `/api/alerts/*` → alert module
  - etc.
- Rate limiting per tenant
- Has no business logic

### Modules expose REST API directly
- Each module has its own HTTP server (not just gRPC)
- Extracts tenant from `X-Tenant` header (injected by load balancer)
- Validates JWT in its own middleware
- Responds directly to frontend

---

## Critical Rules

### ❌ NEVER:
- Import another module
- Logic in `infrastructure/` or `main.go`
- SQL outside `persistence/`
- Hardcode tenant
- Share structs between modules

### ✅ ALWAYS:
- Interfaces in `domain`
- Explicit tenant in all operations
- DTOs in their feature
- Communication via gRPC/Kafka
- Descriptive names

---

## Checklist

- [ ] Do dependencies flow inward?
- [ ] No imports between modules?
- [ ] Tenant in metadata/message?
- [ ] DTOs in their feature?
- [ ] main.go bootstrap only?

---

**When generating code:** Ask if the destination is unclear. Respect dependencies. Treat modules as external.

**No need to create summaries for each implementation you do.**