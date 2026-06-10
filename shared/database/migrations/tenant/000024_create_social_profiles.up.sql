-- Social profile tracking per workspace.
-- platform CHECK constraint validates allowed values at DB level.
-- Unique constraint prevents duplicate (workspace, platform, handle) combos.
-- Partial index on next_check_at speeds up the scheduler's due-profile query.
CREATE TABLE social_profiles (
    id                     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id           UUID        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    platform               TEXT        NOT NULL CHECK (platform IN ('instagram', 'tiktok', 'facebook')),
    handle                 TEXT        NOT NULL,
    display_name           TEXT,
    avatar_url             TEXT,
    is_active              BOOLEAN     NOT NULL DEFAULT TRUE,
    check_interval_minutes INT         NOT NULL DEFAULT 1440,
    next_check_at          TIMESTAMPTZ,
    last_checked_at        TIMESTAMPTZ,
    consecutive_failures   INT         NOT NULL DEFAULT 0,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, platform, handle)
);

-- Partial index: only active profiles are eligible for scheduling.
-- Scheduler WHERE next_check_at <= now() AND is_active = TRUE uses this index.
CREATE INDEX idx_social_profiles_due ON social_profiles (next_check_at)
    WHERE is_active = TRUE;
