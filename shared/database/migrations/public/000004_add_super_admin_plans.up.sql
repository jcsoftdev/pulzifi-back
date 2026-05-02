-- Add plans and organization plan assignments for super-admin company management

CREATE TABLE IF NOT EXISTS public.plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    checks_allowed_monthly INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_plans_code ON public.plans(code);
CREATE INDEX IF NOT EXISTS idx_plans_active ON public.plans(is_active) WHERE is_active = TRUE;

CREATE TABLE IF NOT EXISTS public.organization_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES public.organizations(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES public.plans(id),
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMP NULL,
    created_by UUID NULL REFERENCES public.users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE INDEX IF NOT EXISTS idx_org_plans_org_id ON public.organization_plans(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_plans_plan_id ON public.organization_plans(plan_id);
CREATE INDEX IF NOT EXISTS idx_org_plans_status ON public.organization_plans(status);
CREATE INDEX IF NOT EXISTS idx_org_plans_active ON public.organization_plans(organization_id, started_at DESC)
    WHERE deleted_at IS NULL AND status = 'active';

-- Seed default plans
INSERT INTO public.plans (id, code, name, description, checks_allowed_monthly, is_active) VALUES
    ('20000000-0000-0000-0000-000000000001', 'starter', 'Starter', 'Default plan for new organizations', 1000, TRUE),
    ('20000000-0000-0000-0000-000000000002', 'pro', 'Pro', 'Higher limits for growing teams', 10000, TRUE),
    ('20000000-0000-0000-0000-000000000003', 'enterprise', 'Enterprise', 'Custom enterprise limits', 100000, TRUE)
ON CONFLICT (code) DO NOTHING;

-- Seed SUPER_ADMIN role for platform-level management
INSERT INTO public.roles (id, name, description) VALUES
    ('00000000-0000-0000-0000-000000000004', 'SUPER_ADMIN', 'Platform super administrator with access to all organizations and plans')
ON CONFLICT (name) DO NOTHING;

-- Seed plan management permissions
INSERT INTO public.permissions (id, name, resource, action, description) VALUES
    ('10000000-0000-0000-0000-000000000018', 'plans:read', 'plans', 'read', 'Read plans'),
    ('10000000-0000-0000-0000-000000000019', 'plans:write', 'plans', 'write', 'Create and update plans'),
    ('10000000-0000-0000-0000-000000000020', 'organizations:plan_write', 'organizations', 'plan_write', 'Assign and change organization plans')
ON CONFLICT (name) DO NOTHING;

-- SUPER_ADMIN gets all permissions
INSERT INTO public.role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000004', id FROM public.permissions
ON CONFLICT DO NOTHING;

-- ADMIN can read plans
INSERT INTO public.role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000001', id
FROM public.permissions
WHERE name IN ('plans:read')
ON CONFLICT DO NOTHING;
