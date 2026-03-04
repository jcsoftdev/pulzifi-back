# Usage Module

Billing and usage quota management.

## Use Cases

- `get_metrics` — get usage metrics (checks, pages, workspaces, alerts)
- `get_quotas` — get current billing period quotas
- `list_plans` — list available plans (SUPER_ADMIN)
- `assign_organization_plan` — assign plan to org (SUPER_ADMIN)
- `gift_month` — grant free month to organization (SUPER_ADMIN)

## HTTP Routes (`/usage/*`, tenant-aware)

- GET `/usage/metrics`
- GET `/usage/quotas`
- GET `/usage/admin/plans` (SUPER_ADMIN)
- GET `/usage/admin/organizations` (SUPER_ADMIN)
- PUT `/usage/admin/organizations/{id}/plan` (SUPER_ADMIN)
- POST `/usage/admin/organizations/{id}/gift-month` (SUPER_ADMIN)

## Infrastructure

- PostgreSQL: `usage_tracking`, `organization_plans`, `plans` tables (public/tenant)
- Billing periods anchored to plan start date
- Auto-creates billing periods on first query

## Notes

- Tracks `checks_used` vs `checks_allowed` per billing period
- Supports `storage_period_days` per plan
