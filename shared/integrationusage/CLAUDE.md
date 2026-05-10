# IntegrationUsage Package (`shared/integrationusage/`)

Monthly send-quota enforcement for integration providers (e.g., Twilio SMS).

## Files

- `quota.go` — `Tracker` struct with atomic quota check-and-increment
- `quota_test.go` — Unit tests

## Exported API

### Errors
- `ErrQuotaExceeded` — Returned when org has reached monthly limit

### Types
- `Tracker` — Quota enforcer backed by `integration_send_quotas` table
- `AllowedFunc` — `func(context.Context, uuid.UUID) (int, error)` — resolves allowed count per org

### Functions
- `NewTracker(db *sql.DB) *Tracker`

### Methods (`*Tracker`)
- `CheckAndIncrement(ctx, orgID, serviceType, allowedFor AllowedFunc) error` — Atomically increments `count_used` for the current calendar month. Returns `ErrQuotaExceeded` when limit is reached. Uses PostgreSQL `ON CONFLICT … DO UPDATE … WHERE` for atomicity — no race conditions.

## Table

`integration_send_quotas (org_id, service_type, period_start, count_allowed, count_used)` — unique on `(org_id, service_type, period_start)`.

## Watch Out

- `allowedFor` is called **before** the upsert. If the org has no active plan, `allowedFor` should return `0` to trigger `ErrQuotaExceeded` immediately.
- `period_start` is always the first day of the current UTC month — rows from previous months are never updated.
- The atomic upsert pattern means an empty `RETURNING` (sql.ErrNoRows) signals quota exceeded, not a DB error.
