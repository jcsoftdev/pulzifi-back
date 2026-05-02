ALTER TABLE public.plans ADD COLUMN IF NOT EXISTS storage_period_days INT NOT NULL DEFAULT 7;
UPDATE public.plans SET storage_period_days = 7 WHERE code = 'starter';
UPDATE public.plans SET storage_period_days = 30 WHERE code = 'pro';
UPDATE public.plans SET storage_period_days = 60 WHERE code = 'enterprise';
