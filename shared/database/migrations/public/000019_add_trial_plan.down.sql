-- 000019_add_trial_plan.down.sql
-- Reverse of 000019_add_trial_plan.up.sql. Preserves orphan org_plans rows by deleting
-- the trial plan only if no organization is still referencing it.

DROP INDEX IF EXISTS public.idx_org_plans_trial_ends;

ALTER TABLE public.organization_plans
    DROP COLUMN IF EXISTS converted_at,
    DROP COLUMN IF EXISTS trial_ends_at;

DELETE FROM public.plans
WHERE code = 'trial'
  AND NOT EXISTS (
      SELECT 1 FROM public.organization_plans op WHERE op.plan_id = plans.id
  );
