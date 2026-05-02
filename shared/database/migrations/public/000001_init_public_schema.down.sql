-- Reverse 000001_init_public_schema.up.sql

DROP INDEX IF EXISTS idx_user_oauth_providers_provider;
DROP INDEX IF EXISTS idx_user_oauth_providers_user_id;
DROP TABLE IF EXISTS public.user_oauth_providers;

DROP TABLE IF EXISTS public.registration_requests;

DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP TABLE IF EXISTS public.sessions;

DROP INDEX IF EXISTS idx_org_plans_active;
DROP INDEX IF EXISTS idx_org_plans_status;
DROP INDEX IF EXISTS idx_org_plans_plan_id;
DROP INDEX IF EXISTS idx_org_plans_org_id;
DROP TABLE IF EXISTS public.organization_plans;

DROP INDEX IF EXISTS idx_plans_active;
DROP INDEX IF EXISTS idx_plans_code;
DROP TABLE IF EXISTS public.plans;

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

DROP INDEX IF EXISTS idx_password_resets_token;
DROP TABLE IF EXISTS public.password_resets;

DROP INDEX IF EXISTS idx_refresh_tokens_token;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP TABLE IF EXISTS public.refresh_tokens;

DROP INDEX IF EXISTS idx_organization_members_user_id;
DROP INDEX IF EXISTS idx_organization_members_org_id;
DROP TABLE IF EXISTS public.organization_members;

DROP INDEX IF EXISTS idx_organizations_schema_name;
DROP INDEX IF EXISTS idx_organizations_subdomain;
DROP TABLE IF EXISTS public.organizations;

DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_email_verified;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS public.users;
