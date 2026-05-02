DROP INDEX IF EXISTS idx_orgs_feature_flags;
ALTER TABLE organizations DROP COLUMN IF EXISTS feature_flags;
