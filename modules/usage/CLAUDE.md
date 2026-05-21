# Usage Module

Billing and usage quota management.

## Current Structure

```
modules/usage/
└── infrastructure/
    └── http/
        └── module.go   # All handlers inline (~230 LOC thin orchestration)
```

No `domain/` or `application/` layers exist yet. All logic lives in `infrastructure/http/module.go`.

## HTTP Routes (`/usage/*`, tenant-aware)

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

- All HTTP handlers are inline in `infrastructure/http/module.go` (~230 LOC)
- SUPER_ADMIN role check is done inline via `isSuperAdmin()` helper (not via middleware)
- Tracks `checks_used` vs `checks_allowed` per billing period
- Supports `storage_period_days` per plan

## Watch Out

- Module root still contains legacy `main.go` from the pre-monolith era — can be removed
- No hexagonal structure: domain and application layers were never created

## Architecture Improvements

- **Extract into proper hexagonal layers.** All logic is inline. Create:
  - `domain/entities/`: `Plan`, `UsageTracking`, `BillingPeriod`, `OrganizationPlan`
  - `domain/repositories/`: `PlanRepository`, `UsageRepository`, `OrganizationPlanRepository`
  - `application/`: `get_metrics/`, `get_quotas/`, `list_plans/`, `assign_plan/`, `gift_month/`
  - `infrastructure/persistence/`: PostgreSQL implementations
- **SUPER_ADMIN role check** should use middleware (like admin module) instead of inline checks.
- **Billing period auto-creation** on first query is a side effect in a read operation. Move to a dedicated initialization step or use database triggers.
