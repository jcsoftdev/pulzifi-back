# Integration Module

Third-party service integrations (webhooks, Slack, Teams, etc.).

## Domain Entities

- `Integration` — integration config with service type and flexible config

## Use Cases

- `upsert_integration` — create/update integration
- `list_integrations` — list integrations for org
- `delete_integration` — delete integration
- Webhook CRUD: create, list, get webhook integrations

## HTTP Routes (`/integrations/*`, tenant-aware)

- GET `/integrations`
- POST `/integrations`
- DELETE `/integrations/{id}`
- GET `/integrations/webhooks`
- POST `/integrations/webhooks`
- GET `/integrations/webhooks/{id}`

## Infrastructure

- PostgreSQL: `integrations` table (tenant-scoped)
- Webhook sender: HTTP POST with HMAC signing
- Service types: slack, teams, discord, google_sheets, webhook
