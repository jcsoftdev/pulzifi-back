DROP INDEX IF EXISTS idx_checks_parent_check_id;
ALTER TABLE checks DROP COLUMN IF EXISTS parent_check_id;
