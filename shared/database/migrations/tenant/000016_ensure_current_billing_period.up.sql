-- Ensure a usage_tracking row exists for the current billing period.
--
-- Runs after public migration 000014 which assigns a starter plan to every
-- org that was missing one. Without a current-period row, HasQuota() always
-- returns false and all monitoring checks return 402.
--
-- Billing period is anchored to the day-of-month of the plan's started_at,
-- matching the Go billingPeriodForDate() logic exactly.

DO $$
DECLARE
    v_anchor_day   INT;
    v_checks       INT;
    v_period_start DATE;
    v_period_end   DATE;
    v_next_refill  TIMESTAMP;
    v_today        DATE := CURRENT_DATE;
    v_day          INT  := EXTRACT(DAY  FROM CURRENT_DATE)::INT;
    v_year         INT  := EXTRACT(YEAR FROM CURRENT_DATE)::INT;
    v_month        INT  := EXTRACT(MONTH FROM CURRENT_DATE)::INT;
    v_last_day     INT;
    v_clamped      INT;
    v_prev_year    INT;
    v_prev_month   INT;
    v_nx_year      INT;
    v_nx_month     INT;
BEGIN
    -- Skip if a row already covers today
    IF EXISTS (
        SELECT 1 FROM usage_tracking
        WHERE period_start <= v_today AND period_end >= v_today
    ) THEN
        RETURN;
    END IF;

    -- Look up the org's active plan anchor day and quota
    SELECT
        EXTRACT(DAY FROM op.started_at)::INT,
        p.checks_allowed_monthly
    INTO v_anchor_day, v_checks
    FROM public.organizations o
    JOIN public.organization_plans op ON op.organization_id = o.id
    JOIN public.plans            p  ON p.id = op.plan_id
    WHERE o.schema_name  = current_schema()
      AND op.status      = 'active'
      AND op.deleted_at  IS NULL
    ORDER BY op.started_at DESC
    LIMIT 1;

    IF v_anchor_day IS NULL THEN
        RETURN; -- still no plan, nothing to do
    END IF;

    -- Calculate period_start (matches billingPeriodForDate in Go)
    v_last_day := (date_trunc('month', v_today) + INTERVAL '1 month' - INTERVAL '1 day')::DATE
                  - date_trunc('month', v_today)::DATE + 1;
    v_clamped := LEAST(v_anchor_day, v_last_day);

    IF v_day >= v_clamped THEN
        v_period_start := make_date(v_year, v_month, v_clamped);
    ELSE
        v_prev_month := v_month - 1;
        v_prev_year  := v_year;
        IF v_prev_month < 1 THEN
            v_prev_month := 12;
            v_prev_year  := v_year - 1;
        END IF;
        v_last_day := (make_date(v_prev_year, v_prev_month, 1) + INTERVAL '1 month' - INTERVAL '1 day')::DATE
                      - make_date(v_prev_year, v_prev_month, 1) + 1;
        v_clamped      := LEAST(v_anchor_day, v_last_day);
        v_period_start := make_date(v_prev_year, v_prev_month, v_clamped);
    END IF;

    -- Calculate period_end: day before anchor day of the following month
    v_nx_month := EXTRACT(MONTH FROM v_period_start + INTERVAL '1 month')::INT;
    v_nx_year  := EXTRACT(YEAR  FROM v_period_start + INTERVAL '1 month')::INT;
    v_last_day := (make_date(v_nx_year, v_nx_month, 1) + INTERVAL '1 month' - INTERVAL '1 day')::DATE
                  - make_date(v_nx_year, v_nx_month, 1) + 1;
    v_clamped     := LEAST(v_anchor_day, v_last_day);
    v_period_end  := make_date(v_nx_year, v_nx_month, v_clamped) - INTERVAL '1 day';
    v_next_refill := v_period_end + INTERVAL '1 day';

    INSERT INTO usage_tracking (
        period_start, period_end, checks_allowed, checks_used,
        last_refill_at, next_refill_at, created_at, updated_at
    )
    VALUES (
        v_period_start, v_period_end, v_checks, 0,
        NOW(), v_next_refill, NOW(), NOW()
    )
    ON CONFLICT DO NOTHING;
END $$;
