-- Reverse 000005_seed_default_admin_organization.up.sql

-- Remove role assignments for default user
DELETE FROM public.user_roles
WHERE user_id = (SELECT id FROM public.users WHERE email = 'ajcarlos032@gmail.com')
  AND role_id IN (
    (SELECT id FROM public.roles WHERE name = 'ADMIN'),
    (SELECT id FROM public.roles WHERE name = 'SUPER_ADMIN')
  );

-- Remove default membership
DELETE FROM public.organization_members
WHERE organization_id = (SELECT id FROM public.organizations WHERE subdomain = 'jcsoftdev-inc')
  AND user_id = (SELECT id FROM public.users WHERE email = 'ajcarlos032@gmail.com');

-- Drop tenant schema
DROP SCHEMA IF EXISTS jcsoftdev_inc CASCADE;

-- Remove default organization
DELETE FROM public.organizations WHERE subdomain = 'jcsoftdev-inc';

-- Remove default user
DELETE FROM public.users WHERE email = 'ajcarlos032@gmail.com';
