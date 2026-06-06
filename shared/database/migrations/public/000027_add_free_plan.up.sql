-- 000027_add_free_plan.up.sql
-- Seed the canonical 'free' plan as a real row in public.plans, unifying the
-- previously-implicit "free tier" (no active plan / code outside the paid list)
-- into a first-class plan. Expired trials and orgs without a paid subscription
-- are assigned this plan instead of being left plan-less.
--
-- Limits mirror the pricing doc's Free tier:
--   1 single page, 1 AI insight / month, 3 days storage, email-only alerts.
-- checks_allowed_monthly is not specified by the doc; 50 is a conservative cap
-- for a single monitored page (trial = 100). amount_cents_* are 0 (not NULL) so
-- the CMS renders "$0" instead of "Custom" (formatCents(null) → "Custom").
-- Idempotent so it can be re-applied safely.

INSERT INTO public.plans (
    id, code, name, description,
    checks_allowed_monthly, ai_insights_allowed_monthly, storage_period_days,
    amount_cents_monthly, amount_cents_yearly, currency,
    is_active
)
VALUES (
    '20000000-0000-0000-0000-000000000005',
    'free',
    'Free',
    'Free forever — 1 page, email alerts, limited monthly checks',
    50, 1, 3,
    0, 0, 'usd',
    TRUE
)
ON CONFLICT (code) DO NOTHING;
