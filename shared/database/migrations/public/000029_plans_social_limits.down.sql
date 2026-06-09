-- Rollback: plans_social_limits
-- Scope: public
ALTER TABLE public.plans
    DROP COLUMN IF EXISTS social_profiles_limit,
    DROP COLUMN IF EXISTS social_checks_per_day;
