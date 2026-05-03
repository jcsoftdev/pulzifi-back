CREATE TABLE IF NOT EXISTS org_data_keys (
    org_id        UUID PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE,
    wrapped_dek   BYTEA NOT NULL,
    mek_version   INT NOT NULL DEFAULT 1,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
