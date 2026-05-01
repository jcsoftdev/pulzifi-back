ALTER TABLE public.organizations DROP COLUMN IF EXISTS schema_provisioning_failed_at;

DROP INDEX IF EXISTS public.idx_reg_req_invited_by;
DROP INDEX IF EXISTS public.idx_reg_req_active_invite_email;
DROP INDEX IF EXISTS public.idx_reg_req_invitation_token;

ALTER TABLE public.registration_requests DROP CONSTRAINT IF EXISTS registration_requests_invite_has_expiry;
ALTER TABLE public.registration_requests DROP CONSTRAINT IF EXISTS registration_requests_email_lowercase;

ALTER TABLE public.registration_requests
    DROP COLUMN IF EXISTS accepted_at,
    DROP COLUMN IF EXISTS last_resent_at,
    DROP COLUMN IF EXISTS resent_count,
    DROP COLUMN IF EXISTS revoked_at,
    DROP COLUMN IF EXISTS revoked_by,
    DROP COLUMN IF EXISTS email_error,
    DROP COLUMN IF EXISTS email_sent_at,
    DROP COLUMN IF EXISTS expires_at,
    DROP COLUMN IF EXISTS invited_by,
    DROP COLUMN IF EXISTS invitation_token,
    DROP COLUMN IF EXISTS email;

-- One-way: SET NOT NULL fails if invitation rows exist (intentional).
ALTER TABLE public.registration_requests
    ALTER COLUMN organization_subdomain SET NOT NULL,
    ALTER COLUMN organization_name SET NOT NULL,
    ALTER COLUMN user_id SET NOT NULL;
