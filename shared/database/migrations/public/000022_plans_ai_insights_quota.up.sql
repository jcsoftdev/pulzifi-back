-- Per-plan monthly cap for AI insight generation. NULL = unlimited.
-- Mirrors the existing checks_allowed_monthly column so the plan catalog
-- exposes both limits uniformly. Tenant usage_tracking gets the matching
-- ai_insights_used / ai_insights_allowed counters in tenant migration 000022.
ALTER TABLE public.plans
    ADD COLUMN IF NOT EXISTS ai_insights_allowed_monthly INTEGER NULL;

UPDATE public.plans SET ai_insights_allowed_monthly = 5,    updated_at = NOW() WHERE code = 'trial';
UPDATE public.plans SET ai_insights_allowed_monthly = 50,   updated_at = NOW() WHERE code = 'starter';
UPDATE public.plans SET ai_insights_allowed_monthly = 250,  updated_at = NOW() WHERE code = 'pro';
UPDATE public.plans SET ai_insights_allowed_monthly = NULL, updated_at = NOW() WHERE code = 'enterprise';
