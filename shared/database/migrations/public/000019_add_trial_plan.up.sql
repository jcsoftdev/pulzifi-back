-- 000019_add_trial_plan.up.sql
-- Adds the 'trial' plan row and trial tracking columns on organization_plans.
-- All operations are idempotent so the migration can be re-applied safely.

-- 1. Seed the trial plan. Distinct UUID separate from existing starter/pro/enterprise.
INSERT INTO public.plans (id, code, name, description, checks_allowed_monthly, is_active)
VALUES (
    '20000000-0000-0000-0000-000000000004',
    'trial',
    'Trial',
    '14-day free trial — full access, limited monthly checks',
    500,
    TRUE
)
ON CONFLICT (code) DO NOTHING;

-- 2. Add nullable trial tracking columns to organization_plans (safe — fully backwards-compatible).
ALTER TABLE public.organization_plans
    ADD COLUMN IF NOT EXISTS trial_ends_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS converted_at  TIMESTAMPTZ NULL;

-- 3. Partial index for fast lookup of expired-but-not-converted trials (used by the daily
--    trial-expirer cron and the trial_guard middleware).
CREATE INDEX IF NOT EXISTS idx_org_plans_trial_ends
    ON public.organization_plans (trial_ends_at)
    WHERE trial_ends_at IS NOT NULL
      AND converted_at IS NULL;
