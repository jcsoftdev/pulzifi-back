# Usage Module

## Responsibility

Usage tracking, quota management, plan assignment, and billing metrics. Tracks resource consumption (checks, pages, workspaces, alerts) against plan limits.

## Entities

Uses raw SQL queries (no formal entity structs).

## Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/usage/metrics` | Get usage metrics |
| GET | `/usage/quotas` | Get current quotas and refill date |
| GET | `/usage/admin/plans` | List available plans (SUPER_ADMIN) |
| GET | `/usage/admin/organizations` | List orgs with plans (SUPER_ADMIN) |
| PUT | `/usage/admin/organizations/{id}/plan` | Assign plan to org (SUPER_ADMIN) |

## Dependencies

- Auth middleware
- Organization module
- gRPC server (exposes usage data)

## Constraints

- Tenant-scoped for regular users
- Admin routes require SUPER_ADMIN permission
- Quota refill tracked by billing period
