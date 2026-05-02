-- Ensure default admin organization and tenant schema exist (post-seed fix)

-- Ensure default organization values are present/updated
INSERT INTO public.organizations (name, subdomain, schema_name, owner_user_id)
SELECT
    'jcsoftdev-inc',
    'jcsoftdev-inc',
    'jcsoftdev_inc',
    u.id
FROM public.users u
WHERE u.email = 'ajcarlos032@gmail.com'
ON CONFLICT (subdomain) DO UPDATE
SET
    name = EXCLUDED.name,
    schema_name = EXCLUDED.schema_name,
    owner_user_id = EXCLUDED.owner_user_id,
    updated_at = NOW(),
    deleted_at = NULL;

-- Ensure tenant schema exists so tenant-scope migrations can run
CREATE SCHEMA IF NOT EXISTS jcsoftdev_inc;

-- Ensure default membership exists
INSERT INTO public.organization_members (organization_id, user_id, role)
SELECT o.id, u.id, 'ADMIN'
FROM public.organizations o
JOIN public.users u ON u.email = 'ajcarlos032@gmail.com'
WHERE o.subdomain = 'jcsoftdev-inc'
ON CONFLICT (organization_id, user_id) DO NOTHING;
