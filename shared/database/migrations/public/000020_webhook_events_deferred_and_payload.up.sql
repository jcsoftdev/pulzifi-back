-- Add raw_payload column to stripe_webhook_events so deferred or failed
-- events can be replayed without going back to Stripe.
-- Stores the exact bytes received from Stripe (signature-validated).
ALTER TABLE public.stripe_webhook_events
    ADD COLUMN IF NOT EXISTS raw_payload JSONB NULL;

-- Index supports the reconcile-on-link replay query, which filters by the
-- customer id embedded in the payload and status = 'deferred'.
CREATE INDEX IF NOT EXISTS idx_stripe_webhook_events_status_customer
    ON public.stripe_webhook_events ((raw_payload->'data'->'object'->>'customer'))
    WHERE status = 'deferred';
