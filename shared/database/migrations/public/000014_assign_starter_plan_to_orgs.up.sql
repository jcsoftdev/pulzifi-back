-- Assign the starter plan to all organizations that have no active plan.
--
-- Root cause: organization creation never auto-assigns a plan, so
-- ensureCurrentPeriod cannot create a usage_tracking row, and every
-- quota check returns false (→ 402 Payment Required).
--
-- This backfills existing orgs. Tenant migration 000016 then creates
-- the current billing period for any tenant still missing one.

INSERT INTO public.organization_plans (organization_id, plan_id, status, started_at, created_at, updated_at)
SELECT
    o.id,
    p.id,
    'active',
    NOW(),
    NOW(),
    NOW()
FROM public.organizations o
CROSS JOIN public.plans p
WHERE p.code = 'starter'
  AND p.is_active = TRUE
  AND o.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM public.organization_plans op
      WHERE op.organization_id = o.id
        AND op.status = 'active'
        AND op.deleted_at IS NULL
  )
ON CONFLICT DO NOTHING;
