# Billing Module

Stripe payment gateway — checkout session creation, subscription management via Customer Portal, and webhook-driven plan assignment.

## Key Files

- `infrastructure/http/module.go` — Chi routes; webhook on plain sub-router (no JSON middleware needed), authenticated routes for checkout/portal/subscription
- `infrastructure/persistence/postgres/webhook_event_postgres.go` — Idempotency via `INSERT … ON CONFLICT DO NOTHING` on `stripe_webhook_events`
- `infrastructure/persistence/postgres/subscription_postgres.go` — Reads `organization_plans` + `plans` for subscription state
- `infrastructure/stripe/gateway.go` — Concrete `stripe-go/v79` implementation of `StripeGateway` interface
- `cmd/wiring/billing/plan_assigner.go` — Cross-module adapter that UPSERTs `public.organization_plans` (lives in `cmd/wiring/billing/`, NOT inside this module)

## Hexagonal Layout

```
modules/billing/
├── domain/
│   ├── entities/         — Subscription, Customer, WebhookEvent, BillingStatus
│   ├── repositories/     — SubscriptionRepository, WebhookEventRepository, CustomerRepository (interfaces)
│   └── services/         — StripeGateway, PlanAssigner interfaces + AssignInput struct + ErrPlanNotFound
├── application/
│   ├── create_checkout_session/  — Handler: EnsureCustomer + CreateCheckoutSession
│   ├── create_portal_session/    — Handler: CreatePortalSession
│   ├── get_subscription/         — Handler: reads organization_plans + plans
│   └── handle_webhook/           — Dispatcher: ConstructEvent + idempotency + event sub-handlers
└── infrastructure/
    ├── http/             — module.go (ModuleRegisterer), subscription_handler_test.go
    ├── stripe/           — gateway.go (stripe-go v79 adapter)
    └── persistence/
        ├── postgres/     — webhook_event_postgres.go, subscription_postgres.go, customer_postgres.go
        └── inmem/        — in-memory test doubles (no DB dependency in unit tests)
```

## Webhook Idempotency Strategy

Stripe can deliver the same webhook event more than once. The deduplication contract:

1. `WebhookEventRepository.Save(event)` runs `INSERT INTO public.stripe_webhook_events … ON CONFLICT (event_id) DO NOTHING`.
2. Returns `(true, nil)` on first insert; `(false, nil)` on duplicate — no error, no retry storm.
3. Handler checks the return value: `false` → immediately return HTTP 200 (acknowledged).
4. After full processing: `MarkProcessed(eventID, "processed")` updates `processed_at` and `status`.

This approach survives restarts, multi-instance deployments, and is auditable via direct SQL.

## PlanAssigner Adapter Pattern

`PlanAssigner` is defined as an interface in `modules/billing/domain/services/` but implemented in `cmd/wiring/billing/plan_assigner.go`. This prevents the billing module from importing `modules/usage` or `modules/organization` infrastructure directly (hexagonal boundary enforcement).

The adapter:
- Accepts `AssignInput{OrgID, StripeCustomerID, StripePriceID, BillingStatus, CurrentPeriodEnd, …}`
- Resolves `OrgID` from `stripe_customer_id` when `OrgID == uuid.Nil` (webhook path where only customer ID is known)
- Looks up `plan_id` from `public.plans.stripe_price_id_monthly/yearly`; falls back to `code='starter'` for cancellations
- Deactivates any existing active `organization_plans` row, then inserts a fresh active row
- All operations in a single transaction

`ErrPlanNotFound` is canonical in `domain/services/plan_assigner.go`; `cmd/wiring/billing` re-exports it as `billingwiring.ErrPlanNotFound` for callers that import the wiring package.
`ErrOrphanCustomer` is defined only in `cmd/wiring/billing` (not domain-visible by design).

## BILLING_ENABLED Gate

`cmd/server/modules.go` wraps billing module registration behind `if cfg.BillingEnabled { … }`. With `BILLING_ENABLED=false` (the default):
- No routes are mounted
- No Stripe credentials are read
- No Stripe SDK calls happen

Safe to deploy migrations before flipping the flag.

## Stripe Portal-Only Strategy

This module does NOT use Stripe Elements or custom payment forms. Checkout and subscription management are fully hosted by Stripe:
- **New subscriptions** → `POST /billing/checkout` → redirect to `stripe.com/pay/…`
- **Plan changes / cancellation** → `POST /billing/portal` → redirect to Stripe Customer Portal
- **State sync** → Stripe pushes events to `/billing/webhook`; the handler is the only write-path

This means the frontend never handles raw card data. PCI scope is minimal.

## HTTP Routes

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/billing/checkout` | ADMIN / SUPER_ADMIN | Create Stripe Checkout session |
| POST | `/api/v1/billing/portal` | ADMIN / SUPER_ADMIN | Create Stripe Customer Portal session |
| GET | `/api/v1/billing/subscription` | Any org member | Read current subscription state |
| POST | `/api/v1/billing/webhook` | None (Stripe-Signature) | Receive Stripe webhook events |

## Integration Tests

Opt-in via `//go:build integration`. Require `DATABASE_URL` (or `DB_HOST`) to be set.

```bash
# All billing integration tests:
make test-billing-integration

# Or directly:
DATABASE_URL="postgres://..." \
  go test -tags=integration -v \
    ./modules/billing/infrastructure/persistence/postgres/... \
    ./modules/billing/infrastructure/http/... \
    ./cmd/wiring/billing/...
```

Unit tests (default `go test ./...`) use in-memory repos and mocks — no DB required.

## Watch Out

- **Webhook raw body**: `stripe-go` signs the raw bytes, not the parsed JSON. No JSON middleware must consume `r.Body` before the webhook handler. Verified in `module_test.go` (unit) and `module_integration_test.go` (integration).
- **Public schema only**: all billing tables (`organizations.stripe_customer_id`, `organization_plans`, `stripe_webhook_events`, `plans.stripe_price_id_*`) live in `public` — never `SET search_path` to a tenant schema.
- **JWT roles vs org-member roles**: spec's "owner or admin" maps to `ADMIN` and `SUPER_ADMIN` JWT claims. `organization_members.role` (OWNER/MEMBER) is a separate DB column not present in the JWT.
- **Stripe customer ID persistence**: `CustomerRepository.Save` is an `UPDATE` on `public.organizations.stripe_customer_id` — not a separate customer table.
