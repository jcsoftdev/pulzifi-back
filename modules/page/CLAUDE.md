# Page Module

## Responsibility

Page/URL entity management for monitoring. Pages are URLs registered within workspaces that get periodically checked for changes.

## Entities

- **Page** — ID, WorkspaceID, Name, URL, ThumbnailURL, LastCheckedAt, LastChangeDetectedAt, CheckCount, Tags, CheckFrequency, DetectedChanges

## Repository Interfaces

- `PageRepository` — Create, GetByID, ListByWorkspace, Update, Delete, BulkDelete

## Routes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/pages` | Create page |
| GET | `/pages` | List pages (filterable by workspace) |
| GET | `/pages/{id}` | Get page |
| PUT | `/pages/{id}` | Update page |
| DELETE | `/pages/{id}` | Delete page |
| POST | `/pages/bulk-delete` | Delete multiple pages |

## Dependencies

- Workspace module (pages belong to workspaces)
- Monitoring module (page creation triggers monitoring config)
- gRPC server (exposes page data to other modules)

## Constraints

- Tenant-scoped
- URL must be valid and reachable
- Deleting a page cascades to monitoring configs, checks, insights, and alerts
