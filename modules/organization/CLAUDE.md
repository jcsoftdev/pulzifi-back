# Organization Module

## Responsibility

Multi-tenant organization lifecycle management, subdomain provisioning, and organization membership tracking.

## Entities

- **Organization** — ID, Name, Subdomain, SchemaName, OwnerUserID, CreatedAt, UpdatedAt, DeletedAt

## Repository Interfaces

- `OrganizationRepository` — Create, GetByID, GetBySubdomain, List, Update, Delete, CountBySubdomain

## Routes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/organizations` | Create organization |
| GET | `/organizations` | List organizations |
| GET | `/organizations/{id}` | Get organization |
| PUT | `/organizations/{id}` | Update organization |
| DELETE | `/organizations/{id}` | Delete organization |
| GET | `/organization/current` | Get current org by subdomain |

## Dependencies

- Auth module (user ownership)
- Admin module (triggered on approval)
- gRPC server (exposes organization data to other modules)
- EventBus subscriber (listens for user approval events)

## Constraints

- Organization creation triggers `create_tenant_schema()` SQL function
- Subdomain must be unique across all organizations
- SchemaName derived from subdomain for PostgreSQL schema isolation
