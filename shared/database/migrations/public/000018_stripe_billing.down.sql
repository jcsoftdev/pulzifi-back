-- 000018_stripe_billing.down.sql
-- Drops Stripe billing additions in reverse order.

-- 4. Drop webhook events table
DROP INDEX IF EXISTS idx_stripe_webhook_events_type;
DROP TABLE IF EXISTS stripe_webhook_events;

-- 3. Drop Stripe price columns from plans
ALTER TABLE plans
    DROP COLUMN IF EXISTS stripe_price_id_monthly,
    DROP COLUMN IF EXISTS stripe_price_id_yearly;

-- 2. Drop Stripe subscription columns from organization_plans
ALTER TABLE organization_plans
    DROP COLUMN IF EXISTS stripe_subscription_id,
    DROP COLUMN IF EXISTS stripe_price_id,
    DROP COLUMN IF EXISTS billing_status,
    DROP COLUMN IF EXISTS current_period_end,
    DROP COLUMN IF EXISTS payment_status;

-- 1. Drop Stripe customer column from organizations
ALTER TABLE organizations
    DROP COLUMN IF EXISTS stripe_customer_id;
