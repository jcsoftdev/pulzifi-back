# Monitoring Module

Page monitoring with check scheduling, configuration, sections, and notifications.

## Domain Entities

- `Check` — single page snapshot result (screenshot, HTML, status, change detection)
- `MonitoringConfig` — monitoring configuration per page (frequency, enabled insight types)
- `NotificationPreference` — user notification settings
- `Frequency` — monitoring frequency enum
- `MonitoredSection` — specific section of a page to monitor

## Use Cases (application/ directories)

- `create_check` — create a monitoring check result
- `list_checks` — list checks for a page (with filtering)
- `create_monitoring_config` — set monitoring frequency for a page
- `get_monitoring_config` — get config for a page
- `update_monitoring_config` — update config
- `bulk_update_monitoring_config` — update multiple pages at once
- `create_notification_preference` — set notification settings
- `manage_sections` — CRUD for monitored page sections
- `orchestrator` — background job orchestration (snapshot -> change detection -> insights -> alerts)
- `workers` — concurrent worker pool for check execution

## HTTP Routes (`/monitoring/*`, tenant-aware)

### Checks
- GET `/monitoring/checks/page/{pageId}/stream` — SSE stream for check status (auth only, no tenant required)
- POST `/monitoring/checks` — create check
- GET `/monitoring/checks` — list checks
- GET `/monitoring/checks/{id}` — get check details (inline handler)
- GET `/monitoring/checks/page/{pageId}` — list checks by page

### Configs
- POST `/monitoring/configs` — create monitoring config
- PUT `/monitoring/configs/bulk` — bulk update configs
- GET `/monitoring/configs/{pageId}` — get config by page
- PUT `/monitoring/configs/{pageId}` — update config

### Notification Preferences
- POST `/monitoring/notification-preferences` — create preference
- GET `/monitoring/notification-preferences/{id}` — get preference (inline handler)

### Monitored Sections
- GET `/monitoring/sections/page/{pageId}` — list sections for page
- POST `/monitoring/sections/page/{pageId}` — save sections for page
- DELETE `/monitoring/sections/page/{pageId}/{sectionId}` — delete section

## Infrastructure

- PostgreSQL: `checks`, `monitoring_configs`, `notification_preferences`, `monitored_sections` tables (tenant-scoped)
- Scheduler: cron-based check scheduling (30s poll interval) via `infrastructure/scheduler/`
- Worker Pool: concurrent check execution with failure tracking via `application/workers/`
- Orchestrator: snapshot -> change detection -> insight generation -> alerts via `application/orchestrator/`
- Pub/Sub Broker (CheckBroker): SSE for pushing check completion
- Repository factory pattern: `infrastructure/persistence/repository_factory.go`
- Cross-concern repos: `usage_postgres.go`, `page_postgres.go` in persistence
- Webhook publishing to configured integrations

## Notes

- Most complex module in the codebase (10 application directories)
- Has both API routes and background worker components
- `infrastructure/consumer/` exists but is empty (Kafka consumer placeholder)
- `infrastructure/grpc/` exists for inter-module communication

## Cross-Module Dependencies (violations)

This module directly imports from other modules' infrastructure layers:
- `modules/insight/application/generate_insights` (handler)
- `modules/insight/infrastructure/ai` (OpenRouterGenerator, VisionAnalyzer)
- `modules/snapshot/application` (SnapshotWorker)
- `modules/snapshot/infrastructure/extractor` (HTTPClient)
- `modules/snapshot/infrastructure/storage` (ObjectStorage)
- `modules/email/domain/services` (EmailProvider)

**Recommended:** Define interfaces in this module's domain layer for snapshot capture, insight generation, email sending, and webhook publishing. Inject implementations from `cmd/server/modules.go` to respect hexagonal boundaries.

## Architecture Improvements

- **Orchestrator is tightly coupled.** The orchestrator directly creates insight handlers and snapshot workers. Extract an `Orchestrator` interface in domain and inject dependencies.
- **Repository factory** creates repos dynamically per tenant — consider caching repo instances per tenant to avoid repeated construction.
- **Worker pool** lacks distributed locking — only safe as single-instance. For multi-worker deployments, add PostgreSQL advisory locks or Redis-based distributed locks on `monitoring_config.id`.
- **Scheduler polls at 30s intervals.** For scale, replace with a proper job queue (e.g., pgmq, Redis Streams) to avoid polling overhead.
