DROP INDEX IF EXISTS idx_pages_next_run_at;
ALTER TABLE pages DROP COLUMN IF EXISTS next_run_at;
