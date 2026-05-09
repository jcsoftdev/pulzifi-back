# Feature Flags Package (`shared/featureflags/`)

Per-organization feature flag reader backed by the `organizations.feature_flags` JSONB column.

## Files

- `flags.go` — `Reader` struct with `IsOn` and `Snapshot`
- `flags_test.go` — Unit tests

## Exported API

### Types
- `Reader` — Reads flags from `public.organizations` via the shared `*sql.DB`

### Functions
- `NewReader(db *sql.DB) *Reader`

### Methods (`*Reader`)
- `IsOn(ctx, orgID, key string) (bool, error)` — Returns true only if the dotted-path `key` resolves to boolean `true` in the org's flags blob. Missing org, missing path, or non-bool value all return `(false, nil)` — **default deny**.
- `Snapshot(ctx, orgID) (map[string]any, error)` — Returns the full flags blob for an org. Returns `(nil, nil)` if org not found.

## Key Behavior

- Dot notation for nested flags: `"integrations.twilio"` traverses `{"integrations": {"twilio": true}}`
- Default deny: any missing path, wrong type, or missing org returns false — never panics
- No caching — each call hits the database. For hot paths, consider wrapping with a short-lived local cache.

## Usage

Used by the integration module to gate provider availability per organization (e.g., Twilio is restricted to paid plans).
