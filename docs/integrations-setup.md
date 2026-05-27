# Integrations Phase 1 — Configuration Guide

How to configure the integrations subsystem (Slack + Email + delivery worker) shipped in `feat/integrations-phase-1`.

This file is **gitignored** (`docs/` rule). Distribute via Notion / Confluence / DM, not the repo.

---

## 1. Required environment variables

Add to `.env` (dev) and Dokploy app env (prod).

### Mandatory in production

| Var | Purpose | Format |
|---|---|---|
| `INTEGRATION_TOKEN_KEY` | AES-256-GCM key for at-rest token encryption AND HMAC for OAuth state token. Phase 1 reuses the same key for both. | 32 bytes hex (64 chars). Generate: `openssl rand -hex 32` |
| `INTEGRATION_OAUTH_REDIRECT_BASE` | Root domain that receives OAuth callbacks. The callback path appended is `/api/v1/integrations/oauth/{provider}/callback`. | URL, no trailing slash. Dev: `http://localhost:3000`. Prod: `https://pulzifi.com` |

In dev, if `INTEGRATION_TOKEN_KEY` is unset the server boots with an insecure default `00…00ff` and logs a warning. In production, missing key → fatal startup error (matches `JWT_SECRET` behavior).

### Slack provider

| Var | Purpose |
|---|---|
| `SLACK_CLIENT_ID` | From the Slack app's "Basic Information" page |
| `SLACK_CLIENT_SECRET` | Same |

### Email provider

Reuses the existing `RESEND_API_KEY` / `EMAIL_FROM_ADDRESS` / `EMAIL_FROM_NAME` already in use by other modules (auth, admin, team). No new vars.

### Delivery worker tuning (defaults are fine)

| Var | Default | Purpose |
|---|---|---|
| `DELIVERY_MAX_ATTEMPTS` | `5` | Retries before marking dead |
| `DELIVERY_POLL_INTERVAL` | `5s` | Worker tick rate |
| `DELIVERY_WORKER_POOL_SIZE` | `10` | Concurrent send goroutines per tick |

Backoff schedule (hardcoded): `5s → 30s → 2m → 10m → 1h`, indexed by `attempts-1`, capped at last.

---

## 2. Slack app setup

1. Create app at https://api.slack.com/apps → "From scratch".
2. **OAuth & Permissions** → Bot Token Scopes → add: `chat:write`, `channels:read`, `groups:read`.
3. **OAuth & Permissions** → Redirect URLs → add:
   - Dev: `http://localhost:3000/api/v1/integrations/oauth/slack/callback`
   - Prod: `https://pulzifi.com/api/v1/integrations/oauth/slack/callback`
4. **Basic Information** → copy `Client ID` and `Client Secret` into env.
5. Install app to your dev workspace (button on Basic Information page).

**Note:** the redirect URL lands on the **root** domain, not a tenant subdomain. The signed state token (HMAC) carries the tenant context across the cross-subdomain hop. Don't put the callback under a tenant subdomain — it'll break.

---

## 3. Database migrations

All four already applied during dev:

| Migration | Schema | Purpose |
|---|---|---|
| `000015_create_integrations_oauth.up.sql` | public | `integrations` table (org-level OAuth credentials, encrypted) |
| `000017_create_integration_destinations.up.sql` | tenant | `integration_destinations` table (per-org/workspace/page targets) |
| `000018_create_integration_deliveries.up.sql` | tenant | `integration_deliveries` table (delivery log + retry state) |
| `000019_drop_legacy_integrations.up.sql` | tenant | drops the old tenant-level `integrations` table |

Apply on a fresh env: `make migrate cmd=up`.

The legacy tenant `integrations` table previously held inline `{slack_url, discord_url, ...}` config. **Migration 000019 drops it.** Pre-flight at dev time (Phase 1 plan T8 Step 1) confirmed zero rows. If a prod tenant has rows in that table at deploy time, **stop and reach out** — the migration will lose them. Re-add the row-transformation logic in `000019.up.sql` first.

---

## 4. Architecture overview (so you know what's wired)

```
                    ┌─────────────────────────┐
   change.detected  │  snapshot worker        │
   alert.created    │  (modules/snapshot)     │
        ▲           │  (modules/alert)        │
        │           └──────────┬──────────────┘
        │                      │ PublishDomainEvent
        │                      ▼
        │           ┌─────────────────────────┐
        │           │  EventBus               │
        │           │  (in-memory, node-local)│
        │           └──────────┬──────────────┘
        │                      │ subscribed in cmd/server/modules.go
        │                      ▼
        │           ┌─────────────────────────────┐
        │           │  dispatchevent.Handler      │
        │           │  - resolves destinations    │
        │           │    (scope override:         │
        │           │     page > workspace > org) │
        │           │  - creates 1 Delivery row   │
        │           │    per destination          │
        │           └──────────┬──────────────────┘
        │                      │ INSERT ...status='pending'
        │                      ▼
        │           ┌─────────────────────────────┐
        │           │  integration_deliveries     │
        │           │  (tenant table)             │
        │           └──────────┬──────────────────┘
        │                      │ FOR UPDATE SKIP LOCKED
        │                      ▼
        │           ┌─────────────────────────────┐
        │           │  deliveryworker.Worker      │
        │           │  (cmd/worker — also runs    │
        │           │   in cmd/server when        │
        │           │   ENABLE_WORKERS=true)      │
        │           │  - calls ProviderClient.Send│
        │           │  - MarkDelivered/Failed/Dead│
        │           │  - backoff: 5s→30s→2m→10m→1h│
        │           └──────────┬──────────────────┘
        │                      │
        │             ┌────────┴────────┐
        │             ▼                 ▼
        │        ┌─────────┐      ┌──────────┐
        │        │ Slack   │      │ Email    │
        │        │ chat.   │      │ via      │
        │        │ post-   │      │ Resend   │
        │        │ Message │      │          │
        │        └─────────┘      └──────────┘
        │
        └─── (optionally retried via UI: POST /deliveries/{id}/retry)
```

**Important constraint:** the EventBus is in-memory and **node-local**. If you split `cmd/server` and `cmd/worker` into separate processes, the worker process will NOT receive events published by the server process. For Phase 1:
- **Single process (dev):** `ENABLE_WORKERS=true make dev` → server publishes, in-process worker subscribes. ✅
- **Split processes (prod via Dokploy api + worker apps):** events published in `api` are not seen by `worker`. **Hosts MUST run the snapshot/alert publishers in the same process as the dispatcher subscriber.** The current `cmd/worker` setup uses its own EventBus singleton + dispatcher subscription so the worker-side publishes (snapshot in worker, alert in api) are routed locally. **Cross-process events are dropped.** Phase 4 swaps `MessageBus` to Kafka/Redis pubsub.

---

## 5. HTTP routes added

All under `/api/v1/`. Tenant-scoped (require `X-Tenant` + auth + org membership) **except** the OAuth callback (root-mounted).

| Method | Path | Purpose |
|---|---|---|
| GET | `/integrations` | List org's connected integrations |
| DELETE | `/integrations/{id}` | Disconnect integration + cascade-disable its destinations |
| GET | `/integrations/oauth/{provider}/start` | 302 to provider authorize URL |
| GET | `/integrations/oauth/{provider}/callback` | **Root-mounted, no tenant middleware.** Verifies HMAC state, exchanges code, persists Integration, 302 to tenant return path. |
| GET | `/integrations/{id}/targets` | List Slack channels / etc post-OAuth |
| GET | `/destinations?scope_type=org&scope_id=<uuid>` | List destinations for a scope |
| POST | `/destinations` | Create destination |
| PATCH | `/destinations/{id}` | Patch destination (target/events/enabled) |
| DELETE | `/destinations/{id}` | Hard-delete destination |
| GET | `/deliveries?destination_id=<uuid>&limit=&offset=` | Paginated delivery log |
| POST | `/deliveries/{id}/retry` | Reset a `dead` delivery to `pending` |

---

## 6. Frontend pages

- `/settings/integrations` — provider grid (Slack, Email enabled; Discord/Teams/Sheets/Twilio "coming soon")
- `/settings/integrations/[provider]` — destination form + list + delivery log table

The legacy `/settings` page now just links to `/settings/integrations`. The old paste-a-webhook-URL inputs are gone.

---

## 7. Smoke test (Phase 1 plan T35)

Once Slack creds are in `.env`:

1. `make dev`
2. Browse `http://demo.localhost:3000/settings/integrations` (substitute your dev tenant subdomain).
3. Click **Slack → Connect**. Browser → Slack auth → approve.
4. Returns to `/settings/integrations?connected=slack&toast=success` (green toast).
5. Click Slack card → form loads channels (via `listTargets`) → pick `#alerts` → events: `change.detected` + `alert.created` → save.
6. Trigger a change manually (insert a snapshot diff via `psql` or wait for next scheduled check on a configured page).
7. Within ~10s the Delivery Log shows the row flip `pending → delivered` and Slack receives the formatted message.

**Failure path test:** revoke the Slack bot token from the workspace's app management UI. Trigger another change. Worker hits 401 → marks integration `expired`, delivery `dead` (no retry on auth failures). Both visible in the delivery log.

---

## 8. Operational notes

- **Token refresh:** Phase 1 does NOT actually refresh tokens. The hook is in place (`ProviderClient.RefreshAccessToken`), but Slack bot tokens don't expire so it returns nil. Future providers (Google Sheets, Twilio with rotating tokens) will need real refresh logic.
- **Backoff is wall-clock based.** Restarts re-pick-up rows whose `next_attempt_at` has passed.
- **`FOR UPDATE SKIP LOCKED`** in `ClaimPending` makes parallel worker instances safe — but Phase 1 only runs one worker. Verified in `delivery_postgres_test.go::TestDeliveryRepo_ParallelClaim_NoDuplicates`.
- **Tenant enumeration** happens every poll tick (5s default). At >1000 tenants this becomes a hot query — cache in memory with 60s TTL inside `tickAllTenants` if it bites.
- **Architecture rule violation:** the dispatcher's wiring helpers live in `cmd/wiring/integration/` (not `modules/integration/infrastructure/wiring/` as initially planned). The arch checker rejected the cross-module import `integration → organization` from inside the integration module. The composition root in `cmd/` is the proper home.

---

## 9. Known follow-ups deferred from Phase 1

| Origin | Issue | Where to start |
|---|---|---|
| T9 code-review | `IntegrationPostgresRepository.Update` / `SoftDelete` ignore `RowsAffected`; silent no-op on missing row | `modules/integration/infrastructure/persistence/integration_postgres.go` |
| T9 code-review | `provider_meta` JSON unmarshal errors silently default to `{}`; should surface | same file |
| T9 code-review | Missing `var _ repositories.IntegrationRepository = (*IntegrationPostgresRepository)(nil)` interface guard | same file |
| T20 plan note | `OrgGuard.IsActive` returns `(false, nil)` for inactive org; dispatcher silently no-ops without logging | `cmd/wiring/integration/org_guard.go` + `dispatch_event/handler.go` |
| T26 plan note | Worker tenant query has no caching | `modules/integration/infrastructure/worker/delivery_processor.go::listTenants` |
| T28 design | Snapshot-triggered alerts publish only `change.detected`, not `alert.created`, to avoid double-fan-out. Confirm this is the desired semantics; if destinations should fire on both, add the `alert.created` publish in `snapshot.createAlert`. | `modules/snapshot/application/worker.go::createAlert` |
| T35 | Live smoke test never executed in this branch | this guide §7 |

---

## Phase 2 additions

*Shipped in `feat/integrations-phase-2`. Extends Phase 1 without breaking anything.*

### New environment variables

Add to `.env` (dev) and Dokploy app env (prod) as needed.

#### Discord provider

| Var | Purpose |
|---|---|
| `DISCORD_CLIENT_ID` | From the Discord application's "OAuth2" page |
| `DISCORD_CLIENT_SECRET` | Same |

#### Twilio SMS provider

| Var | Purpose |
|---|---|
| `TWILIO_ACCOUNT_SID` | Platform-owned Twilio account SID (used by paid orgs) |
| `TWILIO_AUTH_TOKEN` | Platform-owned Twilio auth token |
| `TWILIO_FROM_NUMBER` | Platform-owned sender number in E.164 (e.g. `+15551234567`) |
| `TWILIO_PAID_PLANS` | Comma-separated plan codes that use platform Twilio (default: `pro,business`). Enterprise orgs supply BYO credentials via `POST /api/v1/integrations/connect`. |

---

### Discord app setup

1. Create application at https://discord.com/developers/applications → "New Application".
2. Navigate to **OAuth2** → **Redirects** → add:
   - Dev: `http://localhost:3000/api/v1/integrations/oauth/discord/callback`
   - Prod: `https://pulzifi.com/api/v1/integrations/oauth/discord/callback`
3. **OAuth2 → General** → copy `Client ID` and `Client Secret` into env as `DISCORD_CLIENT_ID` / `DISCORD_CLIENT_SECRET`.
4. Under **Bot**, enable the bot and grant **Send Messages** + **Embed Links** permissions.

The callback is root-mounted (no tenant subdomain). The HMAC state token carries the tenant context — same pattern as Slack.

---

### Twilio account setup

1. Sign up or log in at https://console.twilio.com.
2. For a **trial account**: verify caller IDs under **Phone Numbers → Verified Caller IDs** for each number that will receive SMS.
3. From the Twilio Console dashboard, copy:
   - **Account SID** (starts with `AC`)
   - **Auth Token**
4. Purchase or provision a Twilio phone number (the "From" number). Copy it in E.164 format (e.g. `+15551234567`).
5. Set these as `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, `TWILIO_FROM_NUMBER` in your env.

**Enterprise BYO:** enterprise orgs supply their own Twilio credentials via the "Connect Twilio" modal in the UI. These are stored encrypted in the `integrations` table under `provider_meta` (same AES-256-GCM encryption as Slack tokens). The BYO credentials override the platform defaults for that org.

---

### Per-org feature flags

Discord and Twilio provider cards are hidden by default (show "Coming soon") until enabled via the `feature_flags` JSONB column on `organizations`.

**Enable Discord for an org:**
```sql
UPDATE organizations
SET feature_flags = jsonb_set(
  COALESCE(feature_flags, '{}'),
  '{integrations,discord}',
  'true'::jsonb
)
WHERE subdomain = 'demo';
```

**Disable Discord for an org:**
```sql
UPDATE organizations
SET feature_flags = jsonb_set(
  COALESCE(feature_flags, '{}'),
  '{integrations,discord}',
  'false'::jsonb
)
WHERE subdomain = 'demo';
```

**Enable Twilio for an org:**
```sql
UPDATE organizations
SET feature_flags = jsonb_set(
  COALESCE(feature_flags, '{}'),
  '{integrations,twilio}',
  'true'::jsonb
)
WHERE subdomain = 'demo';
```

After changing flags, the integrations panel re-reads `GET /api/v1/auth/me` on next load. No cache bust needed — the panel always fetches fresh.

---

### New HTTP routes (Phase 2)

| Method | Path | Purpose |
|---|---|---|
| GET | `/integrations` | Now filters by `feature_flags` — returns only flag-enabled providers |
| GET | `/integrations/oauth/{provider}/start` | Flag-checked before redirect; returns 403 if provider not enabled for org |
| POST | `/integrations/connect` | BYO credentials for enterprise Twilio: body `{provider, credentials}` |
| GET | `/deliveries/{id}` | Single delivery with full `EventPayload`, `ResponseBody`, `AttemptHistory` |
| POST | `/deliveries/bulk-retry` | Body `{ids: [uuid]}` — returns `{retried, skipped, failed}` |

---

### New database migrations

| Migration | Schema | Purpose |
|---|---|---|
| `000016_add_feature_flags.up.sql` | public | Adds `feature_flags JSONB DEFAULT '{}'` to `organizations` |
| `000020_delivery_attempt_history.up.sql` | tenant | Adds `attempt_history JSONB DEFAULT '[]'` to `integration_deliveries` |

Apply: `make migrate cmd=up` — verify version 16 (public) and 20 (tenant).

---

### Phase 2 smoke test checklist

- [ ] Apply migrations: `make migrate cmd=up` — verify public version = 16, tenant version = 20
- [ ] Check all existing orgs have `feature_flags` populated (at least `{}`):
  ```sql
  SELECT subdomain, feature_flags FROM organizations;
  ```
- [ ] **Discord (flag OFF):** browse `/settings/integrations` — Discord card shows "Coming soon", no Connect button.
- [ ] **Discord (flag ON):** set `{integrations: {discord: true}}` via psql. Refresh page. Card shows "Connect Discord" button.
- [ ] **Discord OAuth flow:** click Connect → Discord auth page → approve → returns to `/settings/integrations?integration=discord&status=connected` with green toast. Channel webhook stored in `integrations` table.
- [ ] **Discord delivery:** trigger a `change.detected` event. Delivery log shows `pending → delivered`. Target Discord channel receives the embed.
- [ ] **Disable Discord flag for one org:** set flag to `false` via psql. Refresh that org's panel — card reverts to "Coming soon". Other orgs with flag ON are unaffected.
- [ ] **Twilio (free plan):** log in as a `starter`-plan org with Twilio flag ON. Card shows disabled "Upgrade to use SMS" button.
- [ ] **Twilio (paid plan):** log in as a `pro`/`business` org. Card shows "Connect SMS (Twilio)". Click → taken directly to destination form. Add phone numbers (E.164). Save. Trigger change. SMS arrives from platform number.
- [ ] **Twilio (enterprise BYO):** log in as an enterprise org. Card shows "Connect SMS (Twilio)". Click → BYO modal opens with 3 inputs. Enter invalid Account SID → backend returns 400 with validation error → modal shows toast, stays open. Enter valid credentials → modal closes, integration row created, page shows "Manage". Trigger change. SMS arrives from BYO number.
- [ ] **E.164 client validation:** in the destination form, enter `0155512345` (no leading +). Submit. Inline error appears: "invalid format". Form does not submit.
- [ ] **Delivery detail drawer:** click any table row (not the checkbox or Retry button). Drawer slides in from the right with full event payload, response body, attempt history timeline.
- [ ] **Bulk select + retry:** trigger 3+ deliveries. Revoke a token to force `dead` status. Check 2 dead rows via checkboxes. Click "Retry selected (2)". Toast: `Retried: 2, Skipped: 0, Failed: 0`. Rows flip to `pending`.
- [ ] **Select all dead:** "Select all" checkbox in header selects only `dead` rows. Non-dead rows have no checkbox.

---

## 10. Disabling the subsystem in an emergency

- Unset `SLACK_CLIENT_ID` / `SLACK_CLIENT_SECRET` → Slack provider still registers but rejects all OAuth callbacks (Slack returns `invalid_client`). Existing destinations keep failing harmlessly until disabled.
- To stop all delivery attempts: stop `cmd/worker` (or set `ENABLE_WORKERS=false` on `cmd/server` if running all-in-one). Pending deliveries pile up in `integration_deliveries` until the worker comes back.
- To roll back the schema: `make migrate cmd=down -steps=4` reverses migrations 000015/000017/000018/000019. **Destructive — drops all integration data.**

---

## Phase 3 additions

*Shipped in `feat/integrations-phase-3`. Extends Phase 2 with Google Sheets, Microsoft Teams, Twilio quota enforcement, and destination scope override.*

### New environment variables

Add to `.env` (dev) and Dokploy app env (prod) as needed.

#### Google Sheets provider

| Var | Purpose |
|---|---|
| `SHEETS_CLIENT_ID` | From the Google Cloud OAuth app's "Credentials" page |
| `SHEETS_CLIENT_SECRET` | Same |

#### Microsoft Teams provider

| Var | Purpose |
|---|---|
| `MICROSOFT_CLIENT_ID` | From the Azure AD app registration |
| `MICROSOFT_CLIENT_SECRET` | Client secret generated under "Certificates & secrets" |

#### Twilio per-org quota

| Var | Default | Purpose |
|---|---|---|
| `TWILIO_QUOTA_PAID_PER_MONTH` | `500` | Maximum SMS deliveries per org per calendar month (UTC). Set to `500` for production. Dev: set low (e.g. `3`) to test quota enforcement. |

---

### Google Cloud OAuth app setup (Sheets)

1. Go to [Google Cloud Console](https://console.cloud.google.com) → select or create a project.
2. **APIs & Services → Library** → enable:
   - **Google Drive API**
   - **Google Sheets API**
3. **APIs & Services → OAuth consent screen**:
   - User type: **External** (for multi-tenant use)
   - Scopes: add `https://www.googleapis.com/auth/spreadsheets` and `https://www.googleapis.com/auth/drive.readonly`
   - Add test users if still in "Testing" status.
4. **APIs & Services → Credentials → Create Credentials → OAuth client ID**:
   - Application type: **Web application**
   - Authorized redirect URIs → add:
     - Dev: `http://localhost:3000/api/v1/integrations/oauth/sheets/callback`
     - Prod: `https://pulzifi.com/api/v1/integrations/oauth/sheets/callback`
5. Copy **Client ID** and **Client Secret** into env as `SHEETS_CLIENT_ID` / `SHEETS_CLIENT_SECRET`.

The callback is root-mounted — same HMAC state pattern as Slack/Discord.

---

### Microsoft Azure AD app setup (Teams)

1. Go to [Azure Portal](https://portal.azure.com) → **Azure Active Directory → App registrations → New registration**.
2. **Supported account types**: select **"Accounts in any organizational directory (Any Azure AD directory - Multitenant)"**.
3. **Redirect URI** (Web) → add:
   - Dev: `http://localhost:3000/api/v1/integrations/oauth/teams/callback`
   - Prod: `https://pulzifi.com/api/v1/integrations/oauth/teams/callback`
4. **API permissions → Add a permission → Microsoft Graph → Delegated permissions**:
   - `ChannelMessage.Send`
   - `Team.ReadBasic.All`
   - `Channel.ReadBasic.All`
   - `offline_access`
5. **Certificates & secrets → New client secret** → copy the secret value (shown once).
6. Copy **Application (client) ID** and the client secret into env as `MICROSOFT_CLIENT_ID` / `MICROSOFT_CLIENT_SECRET`.

**Admin pre-approval:** by default, non-admin Microsoft users are blocked until a tenant admin approves the app. The consent URL pattern is:

```
https://login.microsoftonline.com/common/adminconsent?client_id=<MICROSOFT_CLIENT_ID>&redirect_uri=<INTEGRATION_OAUTH_REDIRECT_BASE>/api/v1/integrations/oauth/teams/callback
```

When a non-admin user tries to connect, the backend redirects to `/settings/integrations?integration_error=consent_required&admin_url=<encoded_consent_url>`. The frontend shows a modal with the URL to copy and forward to the IT admin. Once the admin approves, non-admin users can connect normally.

---

### Twilio quota enforcement

Set `TWILIO_QUOTA_PAID_PER_MONTH=500` (or a lower value for dev/testing) in `.env`. The quota is per-org, per-calendar-month (UTC midnight rollover). When the quota is exceeded:

- The delivery worker marks the delivery `dead` with error message `"monthly SMS quota exceeded"`.
- No further SMS deliveries are attempted for that org until the next calendar month.
- The quota counter is stored in the `integration_usage_quotas` table (public schema, migration 000017).

To test quota enforcement in dev:
```bash
# In .env, set:
TWILIO_QUOTA_PAID_PER_MONTH=3
# Trigger 3 deliveries — all succeed.
# Trigger a 4th — worker marks it dead with quota exceeded message.
```

---

### Per-org feature flags (Sheets + Teams)

Enable Google Sheets and Microsoft Teams for a specific org:

```sql
UPDATE organizations
SET feature_flags = jsonb_set(
  jsonb_set(
    COALESCE(feature_flags, '{}'),
    '{integrations,sheets}',
    'true'::jsonb
  ),
  '{integrations,teams}',
  'true'::jsonb
)
WHERE subdomain = 'demo';
```

Enable individually:

```sql
-- Sheets only
UPDATE organizations
SET feature_flags = jsonb_set(
  COALESCE(feature_flags, '{}'),
  '{integrations,sheets}',
  'true'::jsonb
)
WHERE subdomain = 'demo';

-- Teams only
UPDATE organizations
SET feature_flags = jsonb_set(
  COALESCE(feature_flags, '{}'),
  '{integrations,teams}',
  'true'::jsonb
)
WHERE subdomain = 'demo';
```

---

### New database migrations (Phase 3)

| Migration | Schema | Purpose |
|---|---|---|
| `000017_create_integration_usage_quotas.up.sql` | public | `integration_usage_quotas` table (per-org, per-month, per-service usage counters for quota enforcement) |

Apply: `make migrate cmd=up` — verify public version = 17.

---

### Phase 3 smoke test checklist

- [ ] Apply migration 000017: `make migrate scope=public cmd=up`. Verify table `integration_usage_quotas` exists:
  ```sql
  SELECT * FROM public.integration_usage_quotas LIMIT 5;
  ```
- [ ] **Google Sheets — connect**: enable `{integrations: {sheets: true}}` flag for demo org. Browse `/settings/integrations`. Click **Google Sheets → Connect**. Browser → Google OAuth consent → approve. Returns to integrations page with success toast. `integrations` table has a new `sheets` row.
- [ ] **Google Sheets — pick target**: navigate to `/settings/integrations/sheets`. Spreadsheet and tab dropdowns load from the user's Drive. Select a sheet. Save destination.
- [ ] **Google Sheets — delivery**: trigger a `change.detected` event on a monitored page. Within ~10s, the delivery log shows `pending → delivered`. Open the target Google Sheet — a new row is appended with timestamp, page name, change summary.
- [ ] **Google Sheets — token refresh**: set a short token TTL in the dev OAuth app, or wait 1+ hour. Trigger another change. Worker refreshes the access token via `refresh_token`. Row appended successfully. Check that `provider_meta.access_token` in `integrations` table is updated.
- [ ] **Microsoft Teams — admin user**: enable `{integrations: {teams: true}}` flag. Log in as a user who is a Global Administrator on their Microsoft tenant. Click **Microsoft Teams → Connect**. OAuth flow completes directly. Returns to integrations page with success toast.
- [ ] **Microsoft Teams — non-admin user**: log in as a non-admin Microsoft user in a tenant where the app has NOT been admin-consented. Click Connect. OAuth redirects with `error=consent_required`. Frontend shows the Teams consent modal with the admin approval URL. Copy the URL. Log in as the tenant admin, visit the URL, approve. Non-admin user retries Connect → succeeds.
- [ ] **Microsoft Teams — pick target**: navigate to `/settings/integrations/teams`. Team and channel dropdowns load. Select a channel. Save destination.
- [ ] **Microsoft Teams — delivery**: trigger a `change.detected` event. Delivery log shows `pending → delivered`. Target Teams channel receives the message card.
- [ ] **Twilio quota**: set `TWILIO_QUOTA_PAID_PER_MONTH=3` in dev. Trigger 3 Twilio deliveries — all succeed. Trigger a 4th — delivery worker marks it `dead` with error `"monthly SMS quota exceeded"`. Quota counter in `integration_usage_quotas` shows `3`. On the 1st of next month (UTC), trigger again — succeeds (counter reset).
- [ ] **Destination scope override — workspace**: create a destination scoped to a specific workspace (use the scope picker in the destination form: select "Workspace", choose a workspace). Trigger `change.detected` on a page in that workspace → destination fires. Trigger on a page in a different workspace → destination does NOT fire.
- [ ] **Destination scope override — page**: create a destination scoped to a specific page. Trigger `change.detected` on that exact page → fires. Trigger on another page in the same workspace → does NOT fire.
