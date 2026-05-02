ALTER TABLE public.organization_members
ADD COLUMN IF NOT EXISTS invitation_status VARCHAR(20) NOT NULL DEFAULT 'active';
