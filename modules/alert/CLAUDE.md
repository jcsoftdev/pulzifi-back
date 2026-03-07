# Alert Module

Manages alert notifications for page change detection.

## Domain Entities

- `Alert` — change detection alert with flexible metadata (JSON)

## Use Cases (application/ directories)

- `create_alert` — create a new alert
- `count_unread_alerts` — count unread alerts for a workspace
- `list_all_alerts` — list all alerts (read + unread)
- `mark_all_alerts_read` — mark all alerts as read

## HTTP Routes (`/alerts/*`, tenant-aware)

- POST `/alerts` — create alert
- GET `/alerts` — list alerts (filtered by workspace, inline handler)
- GET `/alerts/unread-count` — count unread alerts
- GET `/alerts/all` — list all alerts
- PUT `/alerts/read-all` — mark all alerts as read
- GET `/alerts/{id}` — get alert details (inline handler)
- PUT `/alerts/{id}` — mark single alert as read (inline handler)
- DELETE `/alerts/{id}` — delete alert (inline handler)

## Infrastructure

- PostgreSQL: `alerts` table (tenant-scoped)

## Notes

- Some handlers (get, update, delete, list by workspace) are implemented inline in module.go rather than as separate use case directories

## Architecture Improvements

### Extract Inline Handlers
Four handlers (get, update, delete, list by workspace) are inline in `module.go`. Extract into dedicated use case directories following the existing pattern (`create_alert`, `count_unread_alerts`, etc.) for consistency and testability.

### Alert Delivery Channels
Currently alerts are stored in the database and exposed via REST. Consider adding:
- Real-time push via SSE (similar to `CheckBroker`/`InsightBroker` pattern)
- Integration with the `integration` module for webhook/Slack/Teams delivery
- Email digest notifications via the `email` module
