DROP INDEX IF EXISTS public.idx_stripe_webhook_events_status_customer;

ALTER TABLE public.stripe_webhook_events
    DROP COLUMN IF EXISTS raw_payload;
