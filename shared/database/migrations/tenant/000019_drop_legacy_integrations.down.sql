-- recreate legacy shape so a rollback can boot, even though we're moving away
CREATE TABLE IF NOT EXISTS integrations (
    id UUID PRIMARY KEY,
    service_type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
