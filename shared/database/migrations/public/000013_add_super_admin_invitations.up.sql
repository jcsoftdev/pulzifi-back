ALTER TABLE public.registration_requests
    ALTER COLUMN user_id DROP NOT NULL,
    ALTER COLUMN organization_name DROP NOT NULL,
    ALTER COLUMN organization_subdomain DROP NOT NULL,
    ADD COLUMN IF NOT EXISTS email VARCHAR(255),
    ADD COLUMN IF NOT EXISTS invitation_token VARCHAR(64),
    ADD COLUMN IF NOT EXISTS invited_by UUID REFERENCES public.users(id),
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS email_sent_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS email_error TEXT,
    ADD COLUMN IF NOT EXISTS revoked_by UUID REFERENCES public.users(id),
    ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS resent_count INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_resent_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS accepted_at TIMESTAMP;

ALTER TABLE public.registration_requests
    ADD CONSTRAINT registration_requests_email_lowercase
    CHECK (email IS NULL OR email = LOWER(email));

CREATE UNIQUE INDEX IF NOT EXISTS idx_reg_req_invitation_token
    ON public.registration_requests(invitation_token)
    WHERE invitation_token IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_reg_req_active_invite_email
    ON public.registration_requests(email)
    WHERE invitation_token IS NOT NULL AND status IN ('pending', 'accepted');

CREATE INDEX IF NOT EXISTS idx_reg_req_invited_by
    ON public.registration_requests(invited_by);

ALTER TABLE public.organizations
    ADD COLUMN IF NOT EXISTS schema_provisioning_failed_at TIMESTAMP;
