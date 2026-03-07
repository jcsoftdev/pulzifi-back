# Integration Module

Third-party service integrations (webhooks, Slack, Teams, etc.).

## Domain Entities

- `Integration` — integration config with service type and flexible config

## Use Cases (application/ directories)

- `upsert_integration` — create/update integration
- `list_integrations` — list integrations for org
- `delete_integration` — delete integration
- `setup_webhook` — webhook setup logic (directory exists but unused by HTTP layer)

## HTTP Routes (`/integrations/*`, tenant-aware)

- GET `/integrations` — list integrations
- POST `/integrations` — create/update integration
- DELETE `/integrations/{id}` — delete integration
- POST `/integrations/webhooks` — create webhook (inline handler)
- GET `/integrations/webhooks` — list webhooks (inline handler)
- GET `/integrations/webhooks/{id}` — get webhook details (inline handler)

## Infrastructure

- PostgreSQL: `integrations` table (tenant-scoped)
- Webhook sender: HTTP POST with HMAC signing
- Service types: slack, teams, discord, google_sheets, webhook

## Notes

- `setup_webhook/` use case directory exists but is not imported by the HTTP module; webhook CRUD is handled inline in module.go

## Architecture Improvements

### Complete `setup_webhook` Use Case
The `setup_webhook/` directory exists but is unused. Either implement the use case and wire it into the HTTP layer, or remove the empty directory to avoid confusion.

### Webhook Reliability
- Add retry logic for failed webhook deliveries (exponential backoff)
- Store webhook delivery history for debugging (delivery status, response codes, timestamps)
- Add webhook signature verification documentation for consumers

### Integration Testing
- Add Slack/Teams/Discord API mocks for testing notification delivery
- Test HMAC signature generation for webhook payloads
