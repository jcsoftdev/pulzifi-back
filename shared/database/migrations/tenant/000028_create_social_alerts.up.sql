-- Social alerts: in-app notification feed for detected social changes and
-- profile suspensions. Separate from the URL-monitoring `alerts` table (which
-- has a NOT NULL FK to check_id) so social rows stay uniform and independently
-- indexed. change_id is nullable because suspension alerts have no change.
CREATE TABLE social_alerts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id   UUID NOT NULL REFERENCES social_profiles(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL,
    change_id    UUID REFERENCES social_changes(id) ON DELETE SET NULL,
    alert_type   VARCHAR(50) NOT NULL,
    title        VARCHAR(255) NOT NULL,
    description  TEXT,
    change_types TEXT[] NOT NULL DEFAULT '{}',
    is_read      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_social_alerts_profile   ON social_alerts (profile_id, created_at DESC);
CREATE INDEX idx_social_alerts_workspace ON social_alerts (workspace_id, created_at DESC);
CREATE INDEX idx_social_alerts_unread    ON social_alerts (workspace_id, is_read) WHERE is_read = FALSE;
