-- Durable audit + recoverability marker for organization deletion.
-- One row per deletion attempt; survives restarts; queryable without grepping logs.
-- No FK on organization_id or actor_id — the org and user rows are hard-deleted
-- during the cascade; the audit row must outlive them.
CREATE TABLE organization_deletions (
    id              UUID        PRIMARY KEY,
    organization_id UUID        NOT NULL,
    schema_name     TEXT        NOT NULL,
    actor_type      TEXT        NOT NULL,
    actor_id        UUID        NULL,
    status          TEXT        NOT NULL DEFAULT 'pending'
                                CHECK (status IN ('pending', 'completed', 'failed')),
    failure_step    TEXT        NULL,
    error_message   TEXT        NULL,
    requested_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at    TIMESTAMPTZ NULL
);

CREATE INDEX idx_org_deletions_org    ON organization_deletions (organization_id);
CREATE INDEX idx_org_deletions_status ON organization_deletions (status);
