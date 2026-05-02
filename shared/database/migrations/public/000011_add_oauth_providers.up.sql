-- Add OAuth providers table for linking external OAuth accounts to users
CREATE TABLE IF NOT EXISTS public.user_oauth_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    access_token TEXT,
    refresh_token TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(provider, provider_user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_oauth_providers_user_id ON public.user_oauth_providers(user_id);
CREATE INDEX IF NOT EXISTS idx_user_oauth_providers_provider ON public.user_oauth_providers(provider, provider_user_id);

-- Add 'used' column to password_resets if not exists
DO $$ BEGIN
    ALTER TABLE public.password_resets ADD COLUMN IF NOT EXISTS used BOOLEAN DEFAULT false;
EXCEPTION WHEN others THEN NULL;
END $$;
