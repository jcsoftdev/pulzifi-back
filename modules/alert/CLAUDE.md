# Alert Module

Manages alert notifications for page change detection.

## Domain Entities

- `Alert` — change detection alert with flexible metadata (JSON)

## Use Cases

- `create_alert` — create a new alert
- `list_alerts` — list alerts filtered by workspace
- `get_alert` — get alert details
- `update_alert` — mark alert as read
- `delete_alert` — delete alert

## HTTP Routes (`/alerts/*`, tenant-aware)

- POST `/alerts`
- GET `/alerts`
- GET `/alerts/{id}`
- PUT `/alerts/{id}`
- DELETE `/alerts/{id}`

## Infrastructure

- PostgreSQL: `alerts` table (tenant-scoped)
