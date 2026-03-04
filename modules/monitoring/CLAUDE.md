# Monitoring Module

Page monitoring with check scheduling, configuration, and notifications.

## Domain Entities

- `Check` — single page snapshot result (screenshot, HTML, status, change detection)
- `MonitoringConfig` — monitoring configuration per page (frequency, enabled insight types)
- `NotificationPreference` — user notification settings
- `Frequency` — monitoring frequency enum

## Use Cases

- `create_check` — create a monitoring check result
- `list_checks` — list checks for a page (with filtering)
- `create_monitoring_config` — set monitoring frequency for a page
- `get_monitoring_config` — get config for a page
- `update_monitoring_config` — update config
- `bulk_update_monitoring_config` — update multiple pages at once
- `create_notification_preference` — set notification settings

## HTTP Routes (`/checks/*`, `/monitoring/*`, tenant-aware)

- POST `/checks`
- GET `/checks`
- POST `/monitoring-config`
- GET `/monitoring-config/{page_id}`
- PUT `/monitoring-config/{page_id}`
- POST `/monitoring-config/bulk-update`

## Infrastructure

- PostgreSQL: `checks`, `monitoring_configs`, `notification_preferences` tables (tenant-scoped)
- Scheduler: cron-based check scheduling (30s poll interval)
- Worker Pool: concurrent check execution with failure tracking
- Pub/Sub Broker (CheckBroker): SSE for pushing check completion
- Snapshot worker, insight generation, email alerts
- Webhook publishing to configured integrations
