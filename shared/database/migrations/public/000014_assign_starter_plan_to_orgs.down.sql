-- Rollback: remove organization_plans rows inserted by migration 000014.
-- Only removes rows that were auto-assigned (no created_by), leaving
-- any manually assigned plans untouched.

DELETE FROM public.organization_plans
WHERE created_by IS NULL
  AND status = 'active'
  AND started_at >= '2026-05-01';
