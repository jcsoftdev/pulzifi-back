# Usage Module

Billing and usage quota management.

## Use Cases (application/ directories)

- `track_usage` — directory exists but is empty (placeholder for future usage tracking implementation)

## HTTP Routes (`/usage/*`, tenant-aware)

All handlers are implemented inline in module.go (no use case directories):

- GET `/usage/metrics` — get usage metrics (checks, pages, workspaces, alerts)
- GET `/usage/quotas` — get current billing period quotas
- GET `/usage/admin/plans` — list available plans (SUPER_ADMIN)
- GET `/usage/admin/organizations` — list organizations with plans (SUPER_ADMIN)
- PUT `/usage/admin/organizations/{id}/plan` — assign plan to org (SUPER_ADMIN)
- POST `/usage/admin/organizations/{id}/gift-month` — grant free month (SUPER_ADMIN)

## Infrastructure

- PostgreSQL: `usage_tracking`, `organization_plans`, `plans` tables (public/tenant)
- Billing periods anchored to plan start date
- Auto-creates billing periods on first query

## Notes

- All HTTP handlers are inline in module.go (~713 lines)
- Domain `entities/` and `repositories/` directories exist but are empty
- Tracks `checks_used` vs `checks_allowed` per billing period
- Supports `storage_period_days` per plan
- `track_usage/` directory is empty (placeholder, no implementation yet)
- SUPER_ADMIN role check is done inline via `isSuperAdmin()` helper (not via middleware)

## Architecture Improvements

- **Extract into proper hexagonal layers.** This module has 713 lines of inline logic. Create:
  - `domain/entities/`: `Plan`, `UsageTracking`, `BillingPeriod`, `OrganizationPlan`
  - `domain/repositories/`: `PlanRepository`, `UsageRepository`, `OrganizationPlanRepository`
  - `application/`: `get_metrics/`, `get_quotas/`, `list_plans/`, `assign_plan/`, `gift_month/`
  - `infrastructure/persistence/`: PostgreSQL implementations
- **SUPER_ADMIN role check** should use middleware (like admin module) instead of inline checks.
- **Billing period auto-creation** on first query is a side effect in a read operation. Move to a dedicated initialization step or use database triggers.
