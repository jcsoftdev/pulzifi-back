-- Social snapshots: one row per check execution.
-- data JSONB stores full ProfileData (bio, followers, posts with stored media URLs).
-- ON DELETE CASCADE means deleting a social_profile cascades to all its snapshots.
CREATE TABLE social_snapshots (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id      UUID        NOT NULL REFERENCES social_profiles(id) ON DELETE CASCADE,
    captured_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    status          TEXT        NOT NULL DEFAULT 'success' CHECK (status IN ('success', 'failed')),
    error           TEXT,
    followers_count INT,
    posts_count     INT,
    data            JSONB
);

CREATE INDEX idx_social_snapshots_profile ON social_snapshots (profile_id, captured_at DESC);

-- Social changes: detected differences between two consecutive snapshots.
-- change_types TEXT[] holds one or more ChangeType enum values.
-- summary JSONB holds structured diff detail (follower delta, bio before/after, new post IDs, etc.).
-- ON DELETE CASCADE means deleting a social_profile cascades to all its changes.
CREATE TABLE social_changes (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id       UUID        NOT NULL REFERENCES social_profiles(id) ON DELETE CASCADE,
    from_snapshot_id UUID        NOT NULL REFERENCES social_snapshots(id),
    to_snapshot_id   UUID        NOT NULL REFERENCES social_snapshots(id),
    change_types     TEXT[]      NOT NULL,
    summary          JSONB       NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_social_changes_profile ON social_changes (profile_id, created_at DESC);
