-- Daily quota tracking for social media checks.
-- Tenant-schema local: one row per usage_date per tenant (no org_id column).
-- Resets naturally each calendar day (new date = new row, old rows accumulate
-- but are only read by today's date key). Mirrors shared/integrationusage pattern.
--
-- Atomic consume pattern (for reference — executed from application code):
--   INSERT INTO social_check_usage (usage_date, checks_used) VALUES (CURRENT_DATE, 1)
--   ON CONFLICT (usage_date) DO UPDATE
--       SET checks_used = social_check_usage.checks_used + 1
--       WHERE social_check_usage.checks_used < $limit
--   RETURNING checks_used;
--   -- No row returned means quota is exhausted.
CREATE TABLE social_check_usage (
    usage_date  DATE PRIMARY KEY,
    checks_used INT  NOT NULL DEFAULT 0
);
