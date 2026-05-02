-- Fix usage_tracking: recalculate billing period based on plan started_at anchor day.
-- Also creates a current-period row if none exists (handles expired periods).

DO $$
DECLARE
    v_anchor_day INT;
    v_checks_allowed INT;
    v_period_start DATE;
    v_period_end DATE;
    v_next_refill TIMESTAMP;
    v_today DATE := CURRENT_DATE;
    v_day INT := EXTRACT(DAY FROM CURRENT_DATE)::INT;
    v_year INT := EXTRACT(YEAR FROM CURRENT_DATE)::INT;
    v_month INT := EXTRACT(MONTH FROM CURRENT_DATE)::INT;
    v_prev_year INT;
    v_prev_month INT;
    v_next_year INT;
    v_next_month INT;
    v_last_day INT;
    v_clamped INT;
BEGIN
    -- Get anchor day and checks from the org's active plan
    SELECT EXTRACT(DAY FROM op.started_at)::INT, p.checks_allowed_monthly
    INTO v_anchor_day, v_checks_allowed
    FROM public.organizations o
    JOIN public.organization_plans op ON op.organization_id = o.id
    JOIN public.plans p ON p.id = op.plan_id
    WHERE o.schema_name = current_schema()
      AND op.deleted_at IS NULL
      AND op.status = 'active'
    ORDER BY op.started_at DESC
    LIMIT 1;

    -- If no plan found, nothing to do
    IF v_anchor_day IS NULL THEN
        RETURN;
    END IF;

    -- Calculate period_start based on anchor day
    v_last_day := (date_trunc('month', v_today) + INTERVAL '1 month' - INTERVAL '1 day')::DATE - date_trunc('month', v_today)::DATE + 1;
    v_clamped := LEAST(v_anchor_day, v_last_day);

    IF v_day >= v_clamped THEN
        -- Period started this month
        v_period_start := make_date(v_year, v_month, v_clamped);
    ELSE
        -- Period started last month
        v_prev_month := v_month - 1;
        v_prev_year := v_year;
        IF v_prev_month < 1 THEN
            v_prev_month := 12;
            v_prev_year := v_year - 1;
        END IF;
        v_last_day := (make_date(v_prev_year, v_prev_month, 1) + INTERVAL '1 month' - INTERVAL '1 day')::DATE - make_date(v_prev_year, v_prev_month, 1) + 1;
        v_clamped := LEAST(v_anchor_day, v_last_day);
        v_period_start := make_date(v_prev_year, v_prev_month, v_clamped);
    END IF;

    -- Calculate period_end: day before the anchor day of the following month
    v_next_month := EXTRACT(MONTH FROM v_period_start + INTERVAL '1 month')::INT;
    v_next_year := EXTRACT(YEAR FROM v_period_start + INTERVAL '1 month')::INT;
    v_last_day := (make_date(v_next_year, v_next_month, 1) + INTERVAL '1 month' - INTERVAL '1 day')::DATE - make_date(v_next_year, v_next_month, 1) + 1;
    v_clamped := LEAST(v_anchor_day, v_last_day);
    v_period_end := make_date(v_next_year, v_next_month, v_clamped) - INTERVAL '1 day';

    v_next_refill := v_period_end + INTERVAL '1 day';

    -- Insert a new period row if one doesn't already cover today
    INSERT INTO usage_tracking (period_start, period_end, checks_allowed, checks_used, last_refill_at, next_refill_at, created_at, updated_at)
    SELECT v_period_start, v_period_end, v_checks_allowed, 0, NOW(), v_next_refill, NOW(), NOW()
    WHERE NOT EXISTS (
        SELECT 1 FROM usage_tracking
        WHERE period_start <= v_today AND period_end >= v_today
    );
END $$;
