CREATE TABLE IF NOT EXISTS integration_deliveries (
    id UUID PRIMARY KEY,
    destination_id UUID NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    event_payload JSONB NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMPTZ,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    response_code INT,
    response_body TEXT,
    error_message TEXT,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_deliveries_pending
    ON integration_deliveries(next_attempt_at) WHERE status='pending';
CREATE INDEX IF NOT EXISTS idx_deliveries_dest_status
    ON integration_deliveries(destination_id, status, created_at DESC);
