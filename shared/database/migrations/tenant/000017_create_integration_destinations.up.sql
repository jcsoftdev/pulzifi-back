CREATE TABLE IF NOT EXISTS integration_destinations (
    id UUID PRIMARY KEY,
    integration_id UUID,
    service_type VARCHAR(32) NOT NULL,
    scope_type VARCHAR(16) NOT NULL CHECK (scope_type IN ('org','workspace','page')),
    scope_id UUID NOT NULL,
    target JSONB NOT NULL,
    events TEXT[] NOT NULL DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_dest_scope
    ON integration_destinations(scope_type, scope_id) WHERE enabled;
CREATE INDEX IF NOT EXISTS idx_dest_events
    ON integration_destinations USING GIN(events);
CREATE INDEX IF NOT EXISTS idx_dest_service_scope
    ON integration_destinations(service_type, scope_type, scope_id) WHERE enabled;
