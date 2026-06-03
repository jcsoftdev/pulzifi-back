-- Transactional outbox: durable event store for at-least-once delivery.
-- Lives in the public schema so cross-tenant atomicity is possible via
-- fully-qualified table references regardless of search_path.

CREATE TABLE IF NOT EXISTS public.outbox_events (
    id              UUID        PRIMARY KEY,
    topic           TEXT        NOT NULL,
    msg_key         TEXT        NOT NULL DEFAULT '',
    payload         JSONB       NOT NULL,
    tenant          TEXT        NOT NULL DEFAULT '',
    status          TEXT        NOT NULL DEFAULT 'pending',   -- pending|publishing|published|dead
    attempts        INT         NOT NULL DEFAULT 0,
    last_error      TEXT        NOT NULL DEFAULT '',
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at    TIMESTAMPTZ
);

-- Relay claim index: find pending rows eligible for dispatch.
CREATE INDEX IF NOT EXISTS idx_outbox_claim
    ON public.outbox_events (next_attempt_at)
    WHERE status = 'pending';

-- Dead-letter monitoring index: inspect failed rows by age.
CREATE INDEX IF NOT EXISTS idx_outbox_dead
    ON public.outbox_events (created_at)
    WHERE status = 'dead';

-- Idempotency table: per-consumer deduplication so redelivery is a no-op.
CREATE TABLE IF NOT EXISTS public.outbox_consumed (
    consumer    TEXT        NOT NULL,
    event_id    UUID        NOT NULL,
    consumed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (consumer, event_id)
);
