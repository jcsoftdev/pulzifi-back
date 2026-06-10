ALTER TABLE public.plans ADD COLUMN IF NOT EXISTS is_popular BOOLEAN NOT NULL DEFAULT FALSE;

-- Enterprise tier has unlimited checks/pages/insights; drop NOT NULL so those cols can be NULL.
ALTER TABLE public.plans ALTER COLUMN checks_allowed_monthly DROP NOT NULL;

UPDATE public.plans SET
    checks_allowed_monthly      = 500,
    max_pages                   = 10,
    max_workspaces              = 3,
    storage_period_days         = 7,
    ai_insights_allowed_monthly = 50,
    is_popular                  = FALSE,
    updated_at                  = NOW()
WHERE code = 'starter';

UPDATE public.plans SET
    checks_allowed_monthly      = 10000,
    max_pages                   = 100,
    max_workspaces              = NULL,
    storage_period_days         = 90,
    ai_insights_allowed_monthly = 250,
    is_popular                  = TRUE,
    updated_at                  = NOW()
WHERE code = 'pro';

UPDATE public.plans SET
    checks_allowed_monthly      = NULL,
    max_pages                   = NULL,
    max_workspaces              = NULL,
    storage_period_days         = 90,
    ai_insights_allowed_monthly = NULL,
    is_popular                  = FALSE,
    updated_at                  = NOW()
WHERE code = 'enterprise';

UPDATE public.plans SET
    is_popular = FALSE,
    updated_at = NOW()
WHERE code IN ('trial', 'free');
