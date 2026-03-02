# Integration Module

## Responsibility

Third-party service integrations including Slack, Microsoft Teams, Discord, Google Sheets, and custom webhooks.

## Entities

- **Integration** — ID, ServiceType (slack/teams/discord/google_sheets/webhook), Config (JSON), Enabled, CreatedBy, CreatedAt, UpdatedAt, DeletedAt

## Repository Interfaces

- `IntegrationRepository` — Create, List, GetByID, GetByServiceType, ListByServiceType, Update, DeleteByID

## Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/integrations` | List integrations |
| POST | `/integrations` | Upsert integration |
| DELETE | `/integrations/{id}` | Delete integration |
| POST | `/integrations/webhooks` | Create webhook |
| GET | `/integrations/webhooks` | List webhooks |
| DELETE | `/integrations/webhooks/{id}` | Delete webhook |

## Dependencies

- Auth middleware
- gRPC server (exposes integration config to monitoring module)

## Constraints

- Tenant-scoped
- Config stored as JSON blob (service-specific structure)
- Webhooks triggered by monitoring events
