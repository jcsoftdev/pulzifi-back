-- Reverse 000002_init_roles_permissions.up.sql

DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP INDEX IF EXISTS idx_user_roles_user_id;
DROP TABLE IF EXISTS public.user_roles;

DROP INDEX IF EXISTS idx_role_permissions_permission_id;
DROP INDEX IF EXISTS idx_role_permissions_role_id;
DROP TABLE IF EXISTS public.role_permissions;

DROP INDEX IF EXISTS idx_permissions_resource_action;
DROP TABLE IF EXISTS public.permissions;

DROP INDEX IF EXISTS idx_roles_name;
DROP TABLE IF EXISTS public.roles;
