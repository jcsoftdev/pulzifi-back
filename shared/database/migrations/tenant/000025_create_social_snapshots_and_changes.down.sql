-- Rollback: create_social_snapshots_and_changes
-- Scope: tenant
-- Drop changes first (FK refs snapshots); then snapshots (FK refs profiles from 000024).
DROP INDEX IF EXISTS idx_social_changes_profile;
DROP TABLE IF EXISTS social_changes;
DROP INDEX IF EXISTS idx_social_snapshots_profile;
DROP TABLE IF EXISTS social_snapshots;
