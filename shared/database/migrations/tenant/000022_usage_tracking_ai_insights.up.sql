-- Tenant counters that mirror checks_used / checks_allowed for AI insight
-- generation. Unlimited (Enterprise) is encoded by writing 2147483647 (INT max)
-- into ai_insights_allowed so the comparison "used < allowed" works uniformly
-- without nullable arithmetic at the application layer.
ALTER TABLE usage_tracking
    ADD COLUMN IF NOT EXISTS ai_insights_used    INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS ai_insights_allowed INTEGER NOT NULL DEFAULT 0;

-- Seed already-active periods so the new cap activates immediately for tenants
-- whose Stripe period_end is still in the future. Pulls the allowed value from
-- the tenant's active organization_plans row joined to public.plans. Tenants
-- whose plan has NULL ai_insights_allowed_monthly (Enterprise) get the
-- INT max sentinel.
UPDATE usage_tracking ut
   SET ai_insights_allowed = COALESCE(p.ai_insights_allowed_monthly, 2147483647)
  FROM public.organization_plans op
  JOIN public.plans         p ON p.id = op.plan_id
  JOIN public.organizations o ON o.id = op.organization_id
 WHERE op.status     = 'active'
   AND op.deleted_at IS NULL
   AND ut.period_end >= CURRENT_DATE
   AND o.schema_name = current_schema();
