-- Seed default admin user and organization for local/dev bootstrap

-- Default user: ajcarlos032@gmail.com / Asdfgh@1
INSERT INTO public.users (email, password_hash, first_name, last_name, email_verified)
VALUES (
    'ajcarlos032@gmail.com',
    '$2a$10$1nlwBIyLx5dYDkl/d3C1leN2sapjLNOI24gUEea532AOJJH7WE7bu',
    'Carlos',
    'Admin',
    TRUE
)
ON CONFLICT (email) DO NOTHING;

-- Default organization
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

-- Ensure tenant schema exists for tenant migrations
CREATE SCHEMA IF NOT EXISTS jcsoftdev_inc;

-- Ensure default membership
INSERT INTO public.organization_members (organization_id, user_id, role)
SELECT o.id, u.id, 'ADMIN'
FROM public.organizations o
JOIN public.users u ON u.email = 'ajcarlos032@gmail.com'
WHERE o.subdomain = 'jcsoftdev-inc'
ON CONFLICT (organization_id, user_id) DO NOTHING;

-- Ensure role assignments for default user
INSERT INTO public.user_roles (user_id, role_id)
SELECT u.id, r.id
FROM public.users u
JOIN public.roles r ON r.name = 'ADMIN'
WHERE u.email = 'ajcarlos032@gmail.com'
ON CONFLICT DO NOTHING;

INSERT INTO public.user_roles (user_id, role_id)
SELECT u.id, r.id
FROM public.users u
JOIN public.roles r ON r.name = 'SUPER_ADMIN'
WHERE u.email = 'ajcarlos032@gmail.com'
ON CONFLICT DO NOTHING;
