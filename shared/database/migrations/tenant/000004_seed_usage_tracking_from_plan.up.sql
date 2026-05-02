-- Seed current active usage period for tenant from organization plan

INSERT INTO usage_tracking (
    period_start,
    period_end,
    checks_allowed,
    checks_used,
    last_refill_at,
    next_refill_at,
    created_at,
    updated_at
)
SELECT
    date_trunc('month', CURRENT_DATE)::date,
    (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month - 1 day')::date,
    COALESCE((
        SELECT p.checks_allowed_monthly
        FROM public.organizations o
        JOIN public.organization_plans op ON op.organization_id = o.id
        JOIN public.plans p ON p.id = op.plan_id
        WHERE o.schema_name = current_schema()
          AND op.deleted_at IS NULL
          AND op.status = 'active'
        ORDER BY op.started_at DESC
        LIMIT 1
    ), 1000),
    0,
    NOW(),
    date_trunc('month', CURRENT_DATE) + INTERVAL '1 month',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1
    FROM usage_tracking ut
    WHERE ut.period_start <= CURRENT_DATE
      AND ut.period_end >= CURRENT_DATE
);
