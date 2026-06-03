ALTER TABLE public.organizations
    ADD COLUMN IF NOT EXISTS company_size            VARCHAR(20)  NULL,
    ADD COLUMN IF NOT EXISTS business_type           VARCHAR(50)  NULL,
    ADD COLUMN IF NOT EXISTS competitor_challenges   JSONB        NULL,
    ADD COLUMN IF NOT EXISTS website_url             TEXT         NULL,
    ADD COLUMN IF NOT EXISTS onboarding_completed_at TIMESTAMPTZ  NULL;
