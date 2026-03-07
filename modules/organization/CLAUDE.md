# Organization Module

Organization entity and multi-tenancy schema management.

## Domain Entities

- `Organization` — organization with subdomain, schema_name, owner

## Use Cases (application/ directories)

- `create_organization` — create org (used by legacy Gin router, NOT by Chi module)
- `get_organization` — fetch org details (used by legacy Gin router, NOT by Chi module)
- `get_current_organization` — fetch authenticated user's org (active, used by Chi module)

## HTTP Routes (`/organizations/*`, `/organization/*`)

- POST `/organizations` — returns 501 Not Implemented (use registration/approval flow instead)
- GET `/organizations` — list organizations (inline handler)
- GET `/organizations/{id}` — get organization details (inline handler)
- PUT `/organizations/{id}` — update organization (inline handler)
- DELETE `/organizations/{id}` — delete organization (inline handler)
- GET `/organization/current` — get current user's organization

## Domain Events

- `organization.created` — published when a new organization is created

## Infrastructure

- PostgreSQL: `organizations` table (public schema)
- gRPC Server: organization service for inter-module communication (proto + generated pb code + server + interceptors)
- Event publishing: `infrastructure/messaging/publisher.go`
- Event subscribing: `infrastructure/messaging/subscriber.go`
- Multi-tenant schema provisioning: `shared/database/migrator.go` `ProvisionTenantSchema()`

## Notes

- Only module with full gRPC implementation (.proto files in `infrastructure/grpc/proto/`)
- Only module with domain events (`domain/events/organization_events.go`)
- Only module with event messaging (publisher/subscriber)
- `infrastructure/http/router.go` contains a legacy Gin-based router (dead code for monolith)
- Schema-per-tenant: each org gets a PostgreSQL schema
- Subdomain routing: tenant extracted from `tenant.app.com`
- E2E tests at module root level: `grpc_e2e_test.go`, `rest_e2e_test.go`
- Stale compiled binary `organization` at module root should be removed

## Architecture Improvements

- **Remove legacy code.** `infrastructure/http/router.go` (Gin router) and `infrastructure/http/handlers/` are dead code from the pre-monolith era. Remove them.
- **Consolidate use cases.** `create_organization` and `get_organization` are only used by the legacy Gin router. Either remove them or wire them into the Chi module.
- **Event-driven provisioning is well-designed.** The `organization.created` event + subscriber pattern is the correct approach. Extend to other modules (e.g., `user.created`, `workspace.created`).
- **gRPC is underutilized.** Only one module has gRPC. Evaluate whether gRPC adds value or if the EventBus pattern is sufficient for inter-module communication.
