# Dashboard Module

## Responsibility

Aggregated dashboard statistics across workspaces, pages, checks, alerts, and insights for the current tenant.

## Entities

- **DashboardStats** — WorkspacesCount, PagesCount, TodayChecksCount, ChangesPerWorkspace[], RecentAlerts[], RecentInsights[]

## Repository Interfaces

- `DashboardRepository` — GetStats

## Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/dashboard/stats` | Get aggregated dashboard statistics |

## Dependencies

- Monitoring, Alert, Insight data (read via cross-table queries within tenant schema)

## Constraints

- Tenant-scoped (stats are per-organization)
- Read-only module (no writes)
