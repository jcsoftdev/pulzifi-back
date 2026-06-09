-- Social media monitoring plan limits.
-- NULL = unlimited (enterprise/trial); 0 = feature disabled for that plan.
-- Mirrors the max_pages / max_workspaces pattern from migration 000028.
-- Idempotent: ADD COLUMN IF NOT EXISTS.
ALTER TABLE public.plans
    ADD COLUMN IF NOT EXISTS social_profiles_limit INT,
    ADD COLUMN IF NOT EXISTS social_checks_per_day INT;

-- Seed values (tune later):
--   free:       1 profile,  2 checks/day  (~every 12h)
--   pro:        10 profiles, 120 checks/day (10 profiles @ every 2h)
--   enterprise: NULL / NULL (unlimited)
--   trial:      NULL / NULL (unlimited during trial)
UPDATE public.plans SET social_profiles_limit = 1,    social_checks_per_day = 2,    updated_at = NOW() WHERE code = 'free';
UPDATE public.plans SET social_profiles_limit = 10,   social_checks_per_day = 120,  updated_at = NOW() WHERE code = 'pro';
UPDATE public.plans SET social_profiles_limit = NULL, social_checks_per_day = NULL, updated_at = NOW() WHERE code = 'enterprise';
UPDATE public.plans SET social_profiles_limit = NULL, social_checks_per_day = NULL, updated_at = NOW() WHERE code = 'trial';
