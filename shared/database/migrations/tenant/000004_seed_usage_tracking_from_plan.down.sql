-- Reverse 000004_seed_usage_tracking_from_plan.up.sql
-- Remove the seeded usage_tracking row for the current period

DELETE FROM usage_tracking
WHERE period_start = date_trunc('month', CURRENT_DATE)::date
  AND period_end = (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month - 1 day')::date;
