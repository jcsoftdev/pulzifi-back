# Integration Module

Multi-provider notification/delivery system: Slack, Discord, Twilio SMS, email, Google Sheets, Microsoft Teams, Gmail, Outlook. Stores encrypted credentials, manages OAuth flows, tracks delivery history, and enforces per-org monthly quotas.

## Key Files

- `infrastructure/http/module.go` — Chi routes for integrations, destinations, deliveries; OAuth callback
- `infrastructure/persistence/integration_postgres.go` — Encrypted credential storage (AES-GCM via `shared/crypto`)
- `infrastructure/persistence/destination_postgres.go` — Delivery destination CRUD
- `infrastructure/persistence/delivery_postgres.go` — Delivery history and retry
- `infrastructure/providers/registry.go` — Registry of all `ProviderClient` implementations
- `domain/services/provider_client.go` — `ProviderClient` interface + `ProviderRegistry`

## Domain Entities

- `Integration` — provider connection with encrypted config (OAuth token or BYO key)
- `Destination` — a named routing target within an integration (e.g., a Slack channel, a phone number)
- `Delivery` — a single notification attempt with status, retry count, error

## Use Cases (application/)

| Directory | What it does |
|-----------|-------------|
| `connect_byo` | Connect a Bring-Your-Own-Key integration (Twilio, webhook) |
| `disconnect_integration` | Remove integration and its destinations |
| `start_oauth` | Build OAuth authorization URL with HMAC-signed state |
| `handle_oauth_callback` | Exchange OAuth code for token, store integration |
| `create_destination` | Create a delivery target for an integration |
| `update_destination` | Rename/reconfigure destination |
| `delete_destination` | Remove destination |
| `list_destinations` | List destinations for an org |
| `list_provider_targets` | Fetch live targets from provider (e.g., Slack channels) |
| `dispatch_event` | Fan-out a `DomainEvent` to all matching destinations |
| `get_delivery` | Get delivery details |
| `list_deliveries` | List deliveries with filtering |
| `retry_delivery` | Retry a failed delivery |
| `bulk_retry_deliveries` | Retry all failed deliveries in a range |

## HTTP Routes (`/integrations`, `/destinations`, `/deliveries` — tenant-aware)

### Integrations
- GET `/integrations` — list connected integrations
- POST `/integrations/connect` — connect BYO (Twilio, etc.)
- DELETE `/integrations/{id}` — disconnect
- GET `/integrations/oauth/{provider}/start` — start OAuth flow
- GET `/integrations/{id}/targets` — list live provider targets
- GET `/api/v1/integrations/oauth/{provider}/callback` — OAuth callback (root-mounted, no tenant middleware)

### Destinations
- GET `/destinations` — list destinations
- POST `/destinations` — create destination
- PATCH `/destinations/{id}` — update destination
- DELETE `/destinations/{id}` — delete destination

### Deliveries
- GET `/deliveries` — list deliveries
- GET `/deliveries/{id}` — get delivery
- POST `/deliveries/bulk-retry` — bulk retry failed deliveries
- POST `/deliveries/{id}/retry` — retry a delivery

## Providers

| Provider | Auth | Key |
|----------|------|-----|
| Slack | OAuth | `SLACK_CLIENT_ID`, `SLACK_CLIENT_SECRET` |
| Discord | OAuth | `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET` |
| Microsoft Teams | OAuth | `MICROSOFT_CLIENT_ID`, `MICROSOFT_CLIENT_SECRET` |
| Gmail | OAuth (reuses Google) | `GMAIL_INTEGRATION_ENABLED=true` + Google creds |
| Outlook | OAuth | `MICROSOFT_CLIENT_ID`, `MICROSOFT_CLIENT_SECRET` |
| Google Sheets | OAuth | `SHEETS_CLIENT_ID`, `SHEETS_CLIENT_SECRET` |
| Twilio SMS | BYO key | Plan-gated; quota via `shared/integrationusage` |
| Email | Adapter wrapping email module | n/a |

## Infrastructure

- `oauth/state_token.go` — HMAC-signed state for OAuth flows (10-min TTL)
- `worker/delivery_processor.go` — Background processor for fan-out and retries
- Encrypted storage: `shared/crypto.AESGCM` wraps OAuth tokens and BYO keys at rest
- Feature flags: `shared/featureflags.Reader` gates Twilio and other restricted providers
- Quota tracking: `shared/integrationusage.Tracker` enforces monthly Twilio SMS limits

## Event Subscription

Subscribes to `TopicChangeDetected` and `TopicAlertCreated` from `shared/eventbus`. On event, `dispatch_event` handler fans out to all destinations that match the event's workspace/page.

## Patterns

- `cmd/wiring/integration/` holds cross-module adapters (email adapter, org plan lookup, quota adapter, tenant repo factory). This is the anti-corruption layer to avoid module-to-module imports.
- OAuth callback is registered directly on the root Chi router (before tenant middleware) because it arrives on the root domain with no subdomain.

## Watch Out

- `INTEGRATION_TOKEN_KEY` must be a valid 32-byte hex string. Defaults to a dev key that is **insecure in production** — logs a warning.
- `Teams` OAuth can return `consent_required` if admin consent is missing. The callback handles this with a special redirect to a consent UI.
- `NewModuleWithDB` is a deprecated shim — always use `NewModule(Deps{...})`.
