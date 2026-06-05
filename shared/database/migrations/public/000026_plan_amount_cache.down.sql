ALTER TABLE public.plans
    DROP COLUMN IF EXISTS amount_cents_monthly,
    DROP COLUMN IF EXISTS amount_cents_yearly,
    DROP COLUMN IF EXISTS currency;
