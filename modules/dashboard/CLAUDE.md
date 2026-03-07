# Dashboard Module

Aggregated dashboard statistics for organization overview.

## Domain Entities

- `DashboardStats` — aggregated metrics (workspaces, pages, checks, alerts, insights)
- `WorkspaceChanges` — detected changes per workspace
- `RecentAlert` — recent alert summary
- `RecentInsight` — recent insight summary

## Use Cases (application/ directories)

- `get_dashboard_stats` — fetch aggregated dashboard statistics

## HTTP Routes (`/dashboard/*`, tenant-aware)

- GET `/dashboard/stats` — get aggregated dashboard statistics

## Infrastructure

- PostgreSQL: complex queries aggregating workspace, page, check, alert, insight data (tenant-scoped)

## Architecture Improvements

### Query Performance
The dashboard aggregates data across multiple tables in a single complex query. For high-traffic orgs:
- Consider materialized views or cached aggregations (Redis) with periodic refresh
- Add database indexes optimized for the dashboard query patterns
- Monitor query execution time and add query timeouts

### Additional Stats
The dashboard could expand to include:
- Trend data (changes over time, chart-ready time series)
- Per-workspace health scores
- SLA compliance metrics
