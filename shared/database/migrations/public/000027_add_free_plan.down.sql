-- 000027_add_free_plan.down.sql
-- Remove the free plan. Guard against orphaning organization_plans rows that
-- reference it: refuse implicitly by leaving the row if any org is on free.
-- DELETE only fires when no organization_plans point at the free plan.
DELETE FROM public.plans p
 WHERE p.code = 'free'
   AND NOT EXISTS (
       SELECT 1 FROM public.organization_plans op WHERE op.plan_id = p.id
   );
