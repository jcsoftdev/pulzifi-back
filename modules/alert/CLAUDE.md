# Alert Module

## Responsibility

Alert creation and management for page change detection events. Alerts notify users when monitored pages have changed.

## Entities

- **Alert** — ID, WorkspaceID, PageID, CheckID, Type, Title, Description, Metadata, ReadAt

## Repository Interfaces

- `AlertRepository` — Create, GetByID, ListByWorkspace, MarkAsRead, Delete

## Routes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/alerts` | Create alert |
| GET | `/alerts` | List alerts (filtered by workspace_id) |
| GET | `/alerts/{id}` | Get alert |
| PUT | `/alerts/{id}` | Mark as read |
| DELETE | `/alerts/{id}` | Delete alert |

## Dependencies

- Monitoring module (triggers alerts on detected changes)
- gRPC server (exposes alert creation to other modules)
- Email module (optional notification)

## Constraints

- Tenant-scoped
- Alerts created automatically when change detection finds differences
- Alerts can be marked as read but not edited
