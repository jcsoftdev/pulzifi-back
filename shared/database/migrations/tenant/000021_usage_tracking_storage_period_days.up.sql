-- usage_tracking.storage_period_days mirrors public.plans.storage_period_days
-- so the billing PlanAssigner can sync the tenant cycle row without a join
-- back into the public schema. Default 7 matches the Starter / Trial plan.
ALTER TABLE usage_tracking
    ADD COLUMN IF NOT EXISTS storage_period_days INTEGER NOT NULL DEFAULT 7;
