-- Rollback: remove usage_tracking rows inserted by migration 000016.
-- Only removes rows where checks_used = 0 (freshly created, no activity yet).

DELETE FROM usage_tracking
WHERE checks_used = 0
  AND created_at >= '2026-05-01';
