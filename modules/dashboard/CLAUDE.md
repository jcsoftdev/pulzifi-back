# Dashboard Module

Aggregated dashboard statistics for organization overview.

## Domain Entities

- `DashboardStats` — aggregated metrics (workspaces, pages, checks, alerts, insights)
- `WorkspaceChanges` — detected changes per workspace
- `RecentAlert` — recent alert summary
- `RecentInsight` — recent insight summary

## Use Cases

- `get_dashboard_stats` — fetch aggregated dashboard statistics

## HTTP Routes (`/dashboard/*`, tenant-aware)

- GET `/dashboard/stats`

## Infrastructure

- PostgreSQL: complex queries aggregating workspace, page, check, alert, insight data (tenant-scoped)
