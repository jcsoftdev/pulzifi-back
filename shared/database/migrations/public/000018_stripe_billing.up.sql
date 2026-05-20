-- 000018_stripe_billing.up.sql
-- Adds Stripe billing columns and event idempotency table (public schema only).
-- All columns are nullable / have safe defaults — fully backwards-compatible.

-- 1. Stripe customer mapping on organizations
ALTER TABLE organizations
    ADD COLUMN stripe_customer_id VARCHAR(255) UNIQUE NULL;

-- 2. Stripe subscription state on organization_plans
ALTER TABLE organization_plans
    ADD COLUMN stripe_subscription_id VARCHAR(255) NULL,
    ADD COLUMN stripe_price_id        VARCHAR(255) NULL,
    ADD COLUMN billing_status         VARCHAR(32)  NOT NULL DEFAULT 'active',
    ADD COLUMN current_period_end     TIMESTAMPTZ  NULL,
    ADD COLUMN payment_status         VARCHAR(32)  NOT NULL DEFAULT 'ok';

-- 3. Stripe price IDs on plans catalog
ALTER TABLE plans
    ADD COLUMN stripe_price_id_monthly VARCHAR(255) NULL,
    ADD COLUMN stripe_price_id_yearly  VARCHAR(255) NULL;

-- 4. Webhook event idempotency log
CREATE TABLE stripe_webhook_events (
    event_id     VARCHAR(255) PRIMARY KEY,
    event_type   VARCHAR(64)  NOT NULL,
    received_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ  NULL,
    status       VARCHAR(32)  NOT NULL DEFAULT 'received'
);

CREATE INDEX idx_stripe_webhook_events_type ON stripe_webhook_events (event_type);
