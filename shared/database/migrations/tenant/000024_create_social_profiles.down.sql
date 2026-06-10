-- Rollback: create_social_profiles
-- Scope: tenant
DROP INDEX IF EXISTS idx_social_profiles_due;
DROP TABLE IF EXISTS social_profiles;
