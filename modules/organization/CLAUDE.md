# Organization Module

Organization entity and multi-tenancy schema management.

## Domain Entities

- `Organization` — organization with subdomain, schema_name, owner

## Use Cases

- `create_organization` — create org (via registration approval)
- `get_organization` — fetch org details
- `get_current_organization` — fetch authenticated user's org

## HTTP Routes (`/organizations/*`, `/organization/*`)

- POST `/organizations`
- GET `/organizations`
- GET `/organizations/{id}`
- PUT `/organizations/{id}`
- DELETE `/organizations/{id}`
- GET `/organization/current`

## Infrastructure

- PostgreSQL: `organizations` table (public schema)
- gRPC Server: organization service for inter-module communication
- Event publishing: `organization.created` event
- Multi-tenant schema creation: `create_tenant_schema()` SQL function

## Notes

- Schema-per-tenant: each org gets a PostgreSQL schema
- Subdomain routing: tenant extracted from `tenant.app.com`
