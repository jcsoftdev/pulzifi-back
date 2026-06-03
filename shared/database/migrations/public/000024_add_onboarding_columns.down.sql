ALTER TABLE public.organizations
    DROP COLUMN IF EXISTS company_size,
    DROP COLUMN IF EXISTS business_type,
    DROP COLUMN IF EXISTS competitor_challenges,
    DROP COLUMN IF EXISTS website_url,
    DROP COLUMN IF EXISTS onboarding_completed_at;
