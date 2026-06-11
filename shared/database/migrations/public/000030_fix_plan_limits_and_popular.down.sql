-- Restore NOT NULL on checks_allowed_monthly before re-adding the constraint.
UPDATE public.plans SET checks_allowed_monthly = 0 WHERE code = 'enterprise' AND checks_allowed_monthly IS NULL;
ALTER TABLE public.plans ALTER COLUMN checks_allowed_monthly SET NOT NULL;

ALTER TABLE public.plans DROP COLUMN IF EXISTS is_popular;
