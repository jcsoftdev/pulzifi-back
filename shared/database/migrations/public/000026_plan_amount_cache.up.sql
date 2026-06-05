-- Cache Stripe price amounts on public.plans so the CMS can render the real
-- charged price without calling Stripe. Written by the billing webhook on
-- price.created / price.updated / product.updated. NULL = no Stripe price
-- (e.g. Enterprise/Custom).
ALTER TABLE public.plans
    ADD COLUMN IF NOT EXISTS amount_cents_monthly INTEGER,
    ADD COLUMN IF NOT EXISTS amount_cents_yearly  INTEGER,
    ADD COLUMN IF NOT EXISTS currency             VARCHAR(3) NOT NULL DEFAULT 'usd';
