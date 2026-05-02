-- Reverse 000004_add_super_admin_plans.up.sql

-- Remove ADMIN's plans:read permission assignment
DELETE FROM public.role_permissions
WHERE role_id = '00000000-0000-0000-0000-000000000001'
  AND permission_id IN (SELECT id FROM public.permissions WHERE name IN ('plans:read'));

-- Remove all SUPER_ADMIN permission assignments
DELETE FROM public.role_permissions
WHERE role_id = '00000000-0000-0000-0000-000000000004';

-- Remove plan management permissions
DELETE FROM public.permissions
WHERE id IN (
    '10000000-0000-0000-0000-000000000018',
    '10000000-0000-0000-0000-000000000019',
    '10000000-0000-0000-0000-000000000020'
);

-- Remove SUPER_ADMIN role
DELETE FROM public.roles
WHERE id = '00000000-0000-0000-0000-000000000004';

-- Remove seeded plans
DELETE FROM public.plans
WHERE id IN (
    '20000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000002',
    '20000000-0000-0000-0000-000000000003'
);

-- Drop tables (reverse order of creation due to FK)
DROP INDEX IF EXISTS idx_org_plans_active;
DROP INDEX IF EXISTS idx_org_plans_status;
DROP INDEX IF EXISTS idx_org_plans_plan_id;
DROP INDEX IF EXISTS idx_org_plans_org_id;
DROP TABLE IF EXISTS public.organization_plans;

DROP INDEX IF EXISTS idx_plans_active;
DROP INDEX IF EXISTS idx_plans_code;
DROP TABLE IF EXISTS public.plans;
