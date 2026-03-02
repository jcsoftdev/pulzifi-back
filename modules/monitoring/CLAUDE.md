# Monitoring Module

## Responsibility

Core monitoring engine: check scheduling, frequency configuration, notification preferences, worker pool management, and real-time check result streaming via SSE.

## Entities

- **Check** — ID, PageID, Status, ScreenshotURL, HTMLSnapshotURL, ContentHash, ChangeDetected, ChangeType, ErrorMessage, DurationMs
- **MonitoringConfig** — ID, PageID, CheckFrequency, ScheduleType, Timezone, BlockAdsCookies, EnabledInsightTypes, EnabledAlertConditions, CustomAlertCondition
- **NotificationPreference** — ID, UserID, AlertType, Enabled

## Repository Interfaces

- `CheckRepository` — Create, GetByID, ListByPage, GetLatestByPage, Update, GetPreviousSuccessfulByPage
- `MonitoringConfigRepository` — Create, GetByPageID, Update, BulkUpdateFrequency, GetDueSnapshotTasks, GetPageURL, UpdateLastCheckedAt, MarkPageDueNow

## Routes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/monitoring/checks` | Create check |
| GET | `/monitoring/checks` | List checks |
| GET | `/monitoring/checks/{id}` | Get check |
| GET | `/monitoring/checks/page/{pageId}` | List checks by page |
| GET | `/monitoring/checks/page/{pageId}/stream` | SSE stream (real-time) |
| POST | `/monitoring/configs` | Create monitoring config |
| GET | `/monitoring/configs/{pageId}` | Get config |
| PUT | `/monitoring/configs/{pageId}` | Update/upsert config |
| PUT | `/monitoring/configs/bulk` | Bulk update frequencies |
| POST | `/monitoring/notification-preferences` | Create preference |
| GET | `/monitoring/notification-preferences/{id}` | Get preference |

## Dependencies

- Snapshot module (worker pool, Playwright extractor)
- Insight module (triggers AI generation on change)
- Email module (notifications)
- Alert module (gRPC client)
- EventBus, CheckBroker (SSE)
- Scheduler (30s poll for due configs)

## Constraints

- Scheduler runs in the worker process (`ENABLE_WORKERS=true`)
- Worker pool manages concurrent goroutines for parallel check execution
- Change detection uses SHA256 hash of extracted HTML content
- SSE uses buffered channels (size 1); slow subscribers get dropped
