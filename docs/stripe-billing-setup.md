# Stripe Billing — Complete Setup Guide

End-to-end Stripe configuration for Pulzifi: from creating the account to going live with a self-serve trial-to-paid funnel.

---

## Table of Contents

1. [Architecture context](#architecture-context)
2. [The `BILLING_ENABLED` flag](#the-billing_enabled-flag)
3. [Trial strategy](#trial-strategy)
4. [Stripe account setup (from scratch)](#stripe-account-setup-from-scratch)
5. [Business profile](#business-profile)
6. [API keys](#api-keys)
7. [Products and prices](#products-and-prices)
8. [Sync price IDs to database](#sync-price-ids-to-database)
9. [Customer Portal](#customer-portal)
10. [Checkout settings](#checkout-settings)
11. [Webhooks](#webhooks)
12. [Test cards](#test-cards)
13. [End-to-end test flow](#end-to-end-test-flow)
14. [Live mode migration](#live-mode-migration)
15. [Recommended live-mode settings](#recommended-live-mode-settings)
16. [Pre-launch checklist](#pre-launch-checklist)
17. [Environment variables](#environment-variables)
18. [Stripe-related code in this repo](#stripe-related-code-in-this-repo)
19. [References](#references)

---

## Architecture context

Pulzifi billing module lives at `modules/billing/` and follows hexagonal architecture.

| Layer | Files |
|---|---|
| **Domain** | `entities/{subscription,customer,webhook_event,billing_status}.go`, `repositories/*.go`, `services/{stripe_gateway,plan_assigner}.go` |
| **Application** | `create_checkout_session/`, `create_portal_session/`, `get_subscription/`, `handle_webhook/` |
| **Infrastructure** | `stripe/gateway.go` (stripe-go v79 adapter), `http/module.go`, `persistence/postgres/*.go`, `persistence/inmem/*.go` |

Key design decisions already implemented:

- **Stripe Portal-only strategy** — frontend never handles card data. All payment UI lives on `stripe.com`. PCI scope is minimal.
- **Webhook idempotency at DB level** — `INSERT … ON CONFLICT DO NOTHING` on `public.stripe_webhook_events`. Duplicate deliveries return 200 immediately.
- **Cross-module adapter** — `PlanAssigner` interface is defined in `modules/billing/domain/services/` but implemented in `cmd/wiring/billing/plan_assigner.go` to avoid importing `modules/usage` or `modules/organization` directly.
- **Migrations** — `shared/database/migrations/public/000018_stripe_billing.up.sql` adds `stripe_customer_id`, `stripe_subscription_id`, `stripe_price_id`, `billing_status`, `current_period_end`, `payment_status`, plus the `stripe_webhook_events` idempotency table. All columns are nullable / have safe defaults — fully backwards-compatible.

---

## The `BILLING_ENABLED` flag

`BILLING_ENABLED` is a runtime feature flag that turns the entire billing module on or off without code changes or redeploys.

### What it controls

In `cmd/server/modules.go`:

```go
if cfg.BillingEnabled {
    billingModule := billinghttp.NewModule(...)
    registry.Register(billingModule)
}
```

When `BILLING_ENABLED=false` (default):
- No `/api/v1/billing/*` routes mounted
- No Stripe credentials read
- No Stripe SDK calls happen
- Webhook endpoint does not exist
- App runs normally without any payment capability

When `BILLING_ENABLED=true`:
- Checkout, portal, subscription, webhook routes active
- Stripe gateway initialized
- Webhook handler subscribed

### Why it exists

| Use case | Benefit |
|---|---|
| Deploy migrations first | Push migration 000018 to prod with billing off, verify schema, then flip the flag |
| Self-hosted / on-prem deploys | Enterprise customers running their own instance don't need Stripe — flag off, app is free |
| Demo / sandbox mode | Pre-launch, app is fully functional without charging anyone |
| Incident response | If Stripe API is down or webhook handling breaks, turn billing off without taking the app down |
| A/B testing | Some tenants with billing, others without |
| Local development | Devs without Stripe keys can still run the app |

It decouples billing infrastructure from core product. Safe to deploy schema changes before flipping the flag.

---

## Trial strategy

Pulzifi uses a **no-card trial** model (Option A) — lower signup friction in exchange for lower trial-to-paid conversion rate.

| | **No-card trial (chosen)** | Card-required trial |
|---|---|---|
| Signup friction | Minimal | High |
| Conversion to paid | 2–5% | 30–60% |
| Signup volume | High | Low |
| Implementation | Custom in our backend | Stripe's `trial_period_days` |
| Best for | Pre-launch growth / PMF discovery | B2B / qualified leads |

### How the trial flow works

```
register → user.status = 'approved' (no manual approval)
        → create organization + provision tenant schema
        → INSERT organization_plans (plan='trial', status='active', trial_ends_at = now + 14d)
        → INSERT usage_tracking current period
        → send welcome email
        → return JWT + redirect to dashboard

day 7  → email "1 week left in trial"
day 12 → email "2 days left — add card"
day 14 → cron marks trial expired
       → middleware blocks POST/PUT/DELETE except /billing/checkout
       → frontend shows upgrade modal (blocking)

upgrade clicked → POST /billing/checkout → Stripe URL → user pays
              → webhook checkout.session.completed
              → PlanAssigner upgrades org to paid plan
              → marks organization_plans.converted_at = now
```

### DB changes for trial

Migration `000019_add_trial_plan.up.sql`:

```sql
INSERT INTO public.plans (id, code, name, description, checks_allowed_monthly, is_active)
VALUES (
    '20000000-0000-0000-0000-000000000000',
    'trial',
    'Trial',
    '14-day free trial — full access, limited monthly checks',
    500,
    TRUE
) ON CONFLICT (code) DO NOTHING;

ALTER TABLE public.organization_plans
    ADD COLUMN IF NOT EXISTS trial_ends_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS converted_at  TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_org_plans_trial_ends
    ON public.organization_plans(trial_ends_at)
    WHERE trial_ends_at IS NOT NULL AND converted_at IS NULL;
```

### Code touch points to implement self-serve + trial

| File | Change |
|---|---|
| `modules/auth/application/register/handler.go` | Create user with `status='approved'`; skip `registration_requests` insert; create org + tenant + trial plan + usage tracking |
| `modules/usage/application/trial_status/` | New use case — GET `/api/v1/usage/trial-status` returns `{ is_trial, days_remaining, trial_ends_at, needs_upgrade }` |
| `cmd/worker/jobs/trial_expiry.go` | New daily cron at 00:00 UTC — finds expired trials, marks `status='trial_expired'`, sends reminder/expiry emails |
| `shared/middleware/trial_guard.go` | New middleware — when trial expired and not converted, allow GET but return 402 on writes (except `/billing/checkout`) |
| `modules/email/templates/` | New templates: `welcome_trial`, `trial_day_7`, `trial_day_12`, `trial_expired`, `trial_converted` |
| `frontend/apps/web/features/billing/ui/` | New `TrialBanner.tsx` + `UpgradeModal.tsx`; show countdown in `AppShell` |
| `modules/billing/application/handle_webhook/handler.go` | On `checkout.session.completed`, set `organization_plans.converted_at = now` |

### Why this matters for valuation

Without self-serve + trial, the app sits behind a manual approval gate that doesn't scale. Adding this flow lifts the valuation from `$50k–$150k` (pre-launch, gated) to `$120k–$250k` (pre-launch, self-serve + trial + Stripe E2E).

---

## Stripe account setup (from scratch)

### Prerequisites

- Permanent email address (avoid `+stripe@` aliases — they cause activation issues later)
- Peruvian RUC or DNI (for live activation)
- USD bank account (Wise, Mercury, Payoneer, or a Peruvian USD account at BCP / Interbank)
- Pulzifi logo in PNG transparent (512×512 minimum)
- Public Terms of Service and Privacy Policy URLs (`pulzifi.com/terms`, `pulzifi.com/privacy`)

### Create the account

1. Open https://dashboard.stripe.com/register
2. Enter email + strong password → "Create account"
3. Verify email via link
4. When prompted for country, select **Peru** (or your country of incorporation)
5. Dashboard opens in **Test mode** (toggle in the top-right shows "Test mode" in orange)
6. **Do not activate live mode yet.** All initial setup happens in test mode.

---

## Business profile

Settings (gear icon, top-right) → **Business settings**

### Public details

- Public business name: `Pulzifi`
- Statement descriptor: `PULZIFI` (max 22 chars — appears on customer's card statement)
- Shortened descriptor: `PULZIFI`
- Customer support email: `support@pulzifi.com`
- Customer support phone: optional
- Public business URL: `https://pulzifi.com`

### Branding

Settings → **Branding**

- Logo: upload transparent PNG
- Icon: favicon 128×128
- Brand color: your primary (e.g. `#6366F1`)
- Accent color: secondary
- Save

---

## API keys

Developers (left menu) → **API keys**

| Key | Format | Where it goes |
|---|---|---|
| Publishable key | `pk_test_51...` | Frontend `.env.local` |
| Secret key | `sk_test_51...` (click "Reveal" to copy) | Backend `.env` (server-only, never committed) |

Save to `.env`:

```bash
STRIPE_SECRET_KEY=sk_test_51...
STRIPE_PUBLISHABLE_KEY=pk_test_51...
BILLING_ENABLED=true
```

For production: consider creating a **Restricted key** (Developers → API keys → "Create restricted key") with only the permissions used: Checkout R/W, Subscriptions R/W, Customers R/W, Webhooks R. Not urgent for test mode.

---

## Products and prices

Catalog (left menu) → **Product catalog** → **Add product**

### Trial — no Stripe product needed

The trial lives only in our DB (`plans.code='trial'`). Stripe does not know about the customer until they upgrade.

### Plan mapping (DB code vs Stripe display name vs pricing)

| DB `plans.code` | Stripe product name (display) | Self-serve in Stripe? | Monthly | Yearly (2 months free = ×10) |
|---|---|---|---|---|
| `starter` | `Pulzifi Starter` | YES | $27 | $270 |
| `pro` | `Pulzifi Professional` | YES | $62 | $620 |
| `enterprise` | — | NO — custom sales pipeline | Custom | — |

The DB keeps `code='pro'` for the middle tier (backend identifier). Only the **Stripe product display name** says "Professional" (matching the landing page UI).

Enterprise is NOT a Stripe product. It uses the "Schedule a Call" sales flow and is assigned manually via the super-admin UI.

### Pulzifi Starter

- Name: `Pulzifi Starter`
- Description: `Perfect for solopreneurs, individual users and business owners. 1 workspace, up to 5 pages, 1 user account, 4 AI insights, 1 week storage, email + messages alerts.`
- Image: upload logo
- More options:
  - Tax code: `txcd_10000000` (Software as a Service)
  - Statement descriptor: `PULZIFI STARTER`

Pricing:
- Pricing model: `Standard pricing`
- Price: `27.00` USD — Billing period: `Monthly`
- Click **Add another price** → Price: `270.00` USD — Billing period: `Yearly` (= 2 months free vs monthly)

Save. Copy both price IDs.

### Pulzifi Professional

- Name: `Pulzifi Professional`
- Description: `Perfect for growing businesses ready to scale. Unlimited workspaces, up to 25 pages, 5 users, unlimited AI insights, multi-channel alerts (Email, Messages, Teams, Slack, Telegram), 1 month storage, priority support.`
- Tax code: `txcd_10000000`
- Statement descriptor: `PULZIFI PRO`

Pricing:
- `62.00` USD monthly
- `620.00` USD yearly (2 months free)

### Pulzifi Enterprise

Do NOT create in Stripe. Enterprise pricing is custom — handled via "Schedule a Call" form on the landing page and assigned manually by super-admin via existing `plans.code='enterprise'` row.

### Final IDs

You should end up with a table like this:

```
starter_monthly = price_1...
starter_yearly  = price_1...
pro_monthly     = price_1...
pro_yearly      = price_1...
```

---

## Sync price IDs to database

Run SQL against your Postgres:

```sql
UPDATE public.plans SET
    stripe_price_id_monthly = 'price_1...starter_monthly',
    stripe_price_id_yearly  = 'price_1...starter_yearly'
WHERE code = 'starter';

UPDATE public.plans SET
    stripe_price_id_monthly = 'price_1...pro_monthly',
    stripe_price_id_yearly  = 'price_1...pro_yearly'
WHERE code = 'pro';

-- Enterprise: NO Stripe IDs — assigned manually via super-admin UI (custom sales)
```

Replace placeholders with your real IDs.

---

## Customer Portal

Settings → **Billing** → **Customer portal** → **Activate**

### Functionality

- Invoice history: ON
- Customer information:
  - Email: ON
  - Billing address: ON
  - Shipping address: OFF (SaaS doesn't ship)
  - Phone number: OFF
  - Tax ID: ON (needed for EU/UK customers)
- Payment methods: add, update, remove — ON
- Cancel subscriptions:
  - Mode: `Cancel at end of billing period`
  - Ask for cancellation reason: ON
  - Reasons: Too expensive / Missing features / Switched to another product / Customer service issues / Don't use enough / Other
- Update subscriptions:
  - Products customers can switch to: Starter, Pro, Enterprise
  - Prorate subscription changes: `Prorate by remaining time`
  - Customers can switch plans: ON
  - Customers can change quantities: OFF

### Business information

- Headline: `Manage your Pulzifi subscription`
- Privacy policy URL: `https://pulzifi.com/privacy`
- Terms of service URL: `https://pulzifi.com/terms`

### Appearance

- Logo: upload
- Brand color: primary
- Accent color: secondary

### Default redirect

- Default return URL: `https://app.pulzifi.com/billing` (production)
- Dev: `http://localhost:3002/billing` (Go is the entry point; it proxies `/billing` to Next.js on 3003 internally)

Save changes.

---

## Checkout settings

Settings → **Checkout and Payment Links**

- Collect customer addresses: ON (needed for tax)
- Enable promotion codes: ON
- Allow customers to adjust quantity: OFF
- Phone number collection: OFF

Custom domain (optional, do later): `Settings → Checkout and Payment Links → Custom domain` to use `pay.pulzifi.com` instead of `checkout.stripe.com/c/...`. Not critical for launch.

---

## Webhooks

### Local development (Stripe CLI)

Install:

```bash
brew install stripe/stripe-cli/stripe
stripe login   # opens browser, authorize
```

Forward webhooks to local backend:

```bash
stripe listen --forward-to localhost:3002/api/v1/billing/webhook
```

The CLI prints a local signing secret (`whsec_local_...`). Use it in your local `.env`:

```bash
STRIPE_WEBHOOK_SECRET=whsec_local_...
```

Keep this command running while developing.

### Production endpoint

Developers → **Webhooks** → **Add endpoint**

- Endpoint URL: `https://app.pulzifi.com/api/v1/billing/webhook`
- Description: `Pulzifi billing sync — production`
- Listen to: **Events on your account**
- Events:

```
checkout.session.completed
checkout.session.expired
customer.created
customer.updated
customer.deleted
customer.subscription.created
customer.subscription.updated
customer.subscription.deleted
customer.subscription.paused
customer.subscription.resumed
customer.subscription.trial_will_end
invoice.created
invoice.finalized
invoice.paid
invoice.payment_failed
invoice.payment_action_required
invoice.upcoming
payment_intent.succeeded
payment_intent.payment_failed
payment_method.attached
payment_method.detached
```

Click **Add endpoint**.

Reveal the signing secret and store it in production `.env`:

```bash
STRIPE_WEBHOOK_SECRET=whsec_live_...
```

Never commit webhook secrets.

---

## Test cards

Use these in test mode:

| Case | Number | CVC | Expiry |
|---|---|---|---|
| Successful payment | `4242 4242 4242 4242` | any 3 digits | any future date |
| Declined | `4000 0000 0000 0002` | `123` | future |
| Requires 3D Secure | `4000 0027 6000 3184` | `123` | future |
| Insufficient funds | `4000 0000 0000 9995` | `123` | future |
| Expired card | `4000 0000 0000 0069` | `123` | future |

Full list: https://docs.stripe.com/testing#cards

---

## End-to-end test flow

### Local setup

Terminal 1 — backend:

```bash
make dev
```

Terminal 2 — Stripe webhook forwarding:

```bash
stripe listen --forward-to localhost:3002/api/v1/billing/webhook
```

### Walkthrough

1. Register a new user in the frontend → should create org + trial plan in DB
2. Verify in DB:

   ```sql
   SELECT op.*, p.code FROM organization_plans op
   JOIN plans p ON p.id = op.plan_id
   WHERE op.organization_id = '<org-id>';
   -- Expect: plan='trial', trial_ends_at = now + 14d
   ```
3. Force trial expiry for quick testing:

   ```sql
   UPDATE organization_plans
   SET trial_ends_at = now() - interval '1 day'
   WHERE organization_id = '<org-id>';
   ```
4. Reload app → blocking upgrade modal should appear
5. Click "Upgrade to Pro Monthly"
6. Frontend calls `POST /api/v1/billing/checkout` with `priceId=price_pro_monthly`
7. Backend returns Stripe Checkout URL
8. Frontend redirects → pay with `4242 4242 4242 4242`
9. Stripe redirects to `/billing?success=true`
10. Stripe CLI terminal shows:

    ```
    checkout.session.completed
    customer.subscription.created
    invoice.paid
    ```
11. Backend webhook processes each → `PlanAssigner` upgrades the org
12. Verify in DB:

    ```sql
    SELECT * FROM organization_plans WHERE organization_id = '<org-id>';
    -- New row: plan_id=pro, status='active', stripe_subscription_id set, converted_at=now
    -- Old row: status='inactive'
    ```
13. Frontend `/billing` shows "Pulzifi Pro — $29/month — Next charge: <date>"

### Manual event triggers

```bash
stripe trigger checkout.session.completed
stripe trigger customer.subscription.created
stripe trigger invoice.payment_failed
stripe trigger customer.subscription.deleted
```

Each should return 200 from the backend.

### Customer Portal test

1. Frontend "Manage billing" button → backend returns portal URL
2. Switch plan Monthly → Yearly → see proration preview → confirm
3. Webhook `customer.subscription.updated` arrives → backend upgrades
4. Cancel subscription → `cancel_at_period_end=true` → access remains until period ends

---

## Live mode migration

### Activate the account

1. Toggle "Test mode" OFF (top-right)
2. Stripe requests activation data:
   - Business type: Individual / Company (SAC, LLC, etc.)
   - Legal name (must match documents)
   - Tax ID / RUC
   - Legal address
   - Industry: `Software`
   - Website: `https://pulzifi.com`
   - Product description: 1–2 paragraphs
   - Avg transaction amount: `$29`
   - Avg monthly volume: estimate
3. Personal details (representative): full name, DOB, DNI/passport, address
4. Bank account for payouts: Wise USD recommended; Peruvian USD account works
5. Submit → Stripe reviews (1–3 days)
6. Activation confirmation email

### Re-create products in live mode

Test mode products and prices do **not** copy to live mode. After activation, toggle to live mode and repeat the [Products and prices](#products-and-prices) section.

Live-mode price IDs will be different:

```
Test:  price_1AbcXyz000001
Live:  price_1MnoPqr000001
```

Update production DB with live IDs.

### Live keys

Developers → API keys (in live mode):

```bash
# Production .env only
STRIPE_SECRET_KEY=sk_live_51...
STRIPE_PUBLISHABLE_KEY=pk_live_51...
```

### Live webhook endpoint

Repeat [Webhooks → Production endpoint](#production-endpoint) in live mode. Get a new `whsec_live_...` and store in production `.env`.

---

## Recommended live-mode settings

### Stripe Tax

Settings → **Tax** → Configure

- Enable Tax
- Origin address: your Peru address
- Default tax behavior: `Exclusive` (price + tax shown separately)
- Activate jurisdictions where you sell (US states, EU countries, UK, AU)
- Cost: 0.5% per transaction where tax applies

### Radar (anti-fraud — free tier)

Active by default. Adjust at `Radar → Rules` if false positives appear.

### Email receipts

Settings → **Emails**:
- Successful payments: ON
- Refunds: ON
- Failed payments: ON
- Customize templates with branding

### Smart Retries

Settings → Subscriptions and emails → **Smart Retries**:
- Enable Smart Retries: ON
- Retry schedule: `8 attempts over 3 weeks`
- After all retries fail: `Cancel subscription`

### Dunning emails

Same section → **Failed payment emails**:
- Send emails on failed payments: ON
- Customer can update payment method via link: ON
- Translate to Spanish if your customers are LatAm

---

## Pre-launch checklist

Before accepting the first real payment:

- [ ] Test mode end-to-end flow works
- [ ] Live webhook endpoint active and returning 200
- [ ] Live products created with price IDs synced to DB
- [ ] `STRIPE_SECRET_KEY=sk_live_...` only in production server env
- [ ] `STRIPE_WEBHOOK_SECRET=whsec_live_...` only in production server env
- [ ] Customer Portal configured in live mode
- [ ] Email templates use `@pulzifi.com` (not Resend default)
- [ ] ToS + Privacy Policy live at `pulzifi.com`
- [ ] Refund policy published
- [ ] Test payout to bank account (Stripe sends $1)
- [ ] Stripe Tax configured if selling internationally
- [ ] Smart Retries enabled
- [ ] Radar rules reviewed
- [ ] Webhook error monitoring (alert if endpoint fails)
- [ ] Reconciliation job for missed webhooks (daily diff between Stripe and DB)

---

## Environment variables

### `.env.development` (local + test mode)

> **Ports note**: This project's local dev ports are NOT the defaults shown in `.env.example` (3000/3001). The script `tools/scripts/assign-dev-ports.sh` rewrites `.env` with values from the port-registry MCP.
>
> Architecture: **Go on 3002** is the single entry point for the user. Next.js runs on **3003** internally, but Go proxies unmatched routes to it — users never hit 3003 directly. All Stripe redirect URLs must point to **3002** (or `pulzifi.lvh.me:3002` when testing tenant subdomains).

```bash
BILLING_ENABLED=true
STRIPE_SECRET_KEY=sk_test_51...
STRIPE_PUBLISHABLE_KEY=pk_test_51...
STRIPE_WEBHOOK_SECRET=whsec_local_...      # from stripe listen
# Only consumed by tools/scripts/setup-stripe-portal.sh; the app builds
# per-request return URLs from the tenant subdomain (/settings/billing).
STRIPE_PORTAL_RETURN_URL=http://localhost:3002/settings/billing

TRIAL_DAYS=14
TRIAL_CHECKS_PER_MONTH=500
TRIAL_EXPIRY_CRON=0 0 * * *
```

Stripe CLI webhook forward (backend on 3002):

```bash
stripe listen --forward-to localhost:3002/api/v1/billing/webhook
```

### `.env.production` (Dokploy, live mode)

```bash
BILLING_ENABLED=true
STRIPE_SECRET_KEY=sk_live_51...
STRIPE_PUBLISHABLE_KEY=pk_live_51...
STRIPE_WEBHOOK_SECRET=whsec_live_...
STRIPE_PORTAL_RETURN_URL=https://app.pulzifi.com/settings/billing

TRIAL_DAYS=14
TRIAL_CHECKS_PER_MONTH=500
TRIAL_EXPIRY_CRON=0 0 * * *
```

---

## Stripe-related code in this repo

### Backend

| Path | Purpose |
|---|---|
| `modules/billing/domain/entities/subscription.go` | `Subscription` entity |
| `modules/billing/domain/entities/customer.go` | `Customer` entity |
| `modules/billing/domain/entities/webhook_event.go` | `WebhookEvent` for idempotency log |
| `modules/billing/domain/entities/billing_status.go` | `BillingStatus` enum |
| `modules/billing/domain/services/stripe_gateway.go` | `StripeGateway` interface |
| `modules/billing/domain/services/plan_assigner.go` | `PlanAssigner` interface + `AssignInput` + `ErrPlanNotFound` |
| `modules/billing/domain/repositories/*.go` | `CustomerRepository`, `SubscriptionRepository`, `WebhookEventRepository` interfaces |
| `modules/billing/application/create_checkout_session/` | Use case: ensure customer + create Stripe Checkout session |
| `modules/billing/application/create_portal_session/` | Use case: create Customer Portal session |
| `modules/billing/application/get_subscription/` | Use case: read current subscription state |
| `modules/billing/application/handle_webhook/` | Webhook dispatcher: signature verify + idempotency + per-event handlers |
| `modules/billing/infrastructure/http/module.go` | Chi routes — webhook on plain sub-router, authenticated routes for the rest |
| `modules/billing/infrastructure/stripe/gateway.go` | `stripe-go/v79` implementation of `StripeGateway` |
| `modules/billing/infrastructure/persistence/postgres/webhook_event_postgres.go` | Idempotent insert via `ON CONFLICT DO NOTHING` |
| `modules/billing/infrastructure/persistence/postgres/subscription_postgres.go` | Reads `organization_plans` + `plans` for state |
| `cmd/wiring/billing/plan_assigner.go` | Cross-module adapter — resolves org from Stripe customer ID and upserts `organization_plans` |
| `shared/database/migrations/public/000018_stripe_billing.up.sql` | Schema changes — backwards-compatible |

### Frontend

| Path | Purpose |
|---|---|
| `frontend/packages/services/src/billing-api.ts` | API client for billing endpoints |
| `frontend/apps/web/features/billing/domain/plan.ts` | `Plan` type |
| `frontend/apps/web/features/billing/domain/subscription.ts` | `Subscription` type |
| `frontend/apps/web/features/billing/application/useCreateCheckout.ts` | Hook: POST checkout, redirect to Stripe URL |
| `frontend/apps/web/features/billing/application/useCustomerPortal.ts` | Hook: POST portal, redirect to Stripe URL |
| `frontend/apps/web/features/billing/application/useSubscription.ts` | Hook: GET current subscription |
| `frontend/apps/web/features/billing/ui/BillingTab.tsx` | Main billing page |
| `frontend/apps/web/features/billing/ui/ManageBillingButton.tsx` | Opens Customer Portal |
| `frontend/apps/web/features/billing/ui/PlanCard.tsx` | Plan selection card |
| `frontend/apps/web/features/billing/ui/SubscriptionStatusCard.tsx` | Active subscription display |

### HTTP routes

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/billing/checkout` | ADMIN / SUPER_ADMIN | Create Stripe Checkout session |
| POST | `/api/v1/billing/portal` | ADMIN / SUPER_ADMIN | Create Customer Portal session |
| GET | `/api/v1/billing/subscription` | Any org member | Read current subscription state |
| POST | `/api/v1/billing/webhook` | None (Stripe-Signature header) | Receive Stripe webhook events |

---

## References

- Stripe Checkout — https://docs.stripe.com/payments/checkout
- Customer Portal — https://docs.stripe.com/customer-management
- Webhooks — https://docs.stripe.com/webhooks
- Test cards — https://docs.stripe.com/testing
- Subscriptions overview — https://docs.stripe.com/billing/subscriptions/overview
- Stripe CLI — https://docs.stripe.com/stripe-cli
- stripe-go SDK — https://github.com/stripe/stripe-go
