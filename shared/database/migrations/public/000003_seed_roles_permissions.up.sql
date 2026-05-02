-- Initial seed data for roles and permissions

-- Insert default roles
INSERT INTO public.roles (id, name, description) VALUES
    ('00000000-0000-0000-0000-000000000001', 'ADMIN', 'Administrator with full access'),
    ('00000000-0000-0000-0000-000000000002', 'USER', 'Standard user with limited access'),
    ('00000000-0000-0000-0000-000000000003', 'VIEWER', 'Read-only access')
ON CONFLICT (name) DO NOTHING;

-- Insert default permissions
INSERT INTO public.permissions (id, name, resource, action, description) VALUES
    ('10000000-0000-0000-0000-000000000001', 'workspaces:read', 'workspaces', 'read', 'Read workspaces'),
    ('10000000-0000-0000-0000-000000000002', 'workspaces:write', 'workspaces', 'write', 'Create and update workspaces'),
    ('10000000-0000-0000-0000-000000000003', 'workspaces:delete', 'workspaces', 'delete', 'Delete workspaces'),
    ('10000000-0000-0000-0000-000000000004', 'pages:read', 'pages', 'read', 'Read pages'),
    ('10000000-0000-0000-0000-000000000005', 'pages:write', 'pages', 'write', 'Create and update pages'),
    ('10000000-0000-0000-0000-000000000006', 'pages:delete', 'pages', 'delete', 'Delete pages'),
    ('10000000-0000-0000-0000-000000000007', 'monitoring:read', 'monitoring', 'read', 'Read monitoring data'),
    ('10000000-0000-0000-0000-000000000008', 'monitoring:write', 'monitoring', 'write', 'Configure monitoring'),
    ('10000000-0000-0000-0000-000000000009', 'alerts:read', 'alerts', 'read', 'Read alerts'),
    ('10000000-0000-0000-0000-000000000010', 'alerts:write', 'alerts', 'write', 'Create and update alerts'),
    ('10000000-0000-0000-0000-000000000011', 'reports:read', 'reports', 'read', 'Read reports'),
    ('10000000-0000-0000-0000-000000000012', 'reports:write', 'reports', 'write', 'Generate reports'),
    ('10000000-0000-0000-0000-000000000013', 'users:read', 'users', 'read', 'Read users'),
    ('10000000-0000-0000-0000-000000000014', 'users:write', 'users', 'write', 'Create and update users'),
    ('10000000-0000-0000-0000-000000000015', 'users:delete', 'users', 'delete', 'Delete users'),
    ('10000000-0000-0000-0000-000000000016', 'organizations:read', 'organizations', 'read', 'Read organizations'),
    ('10000000-0000-0000-0000-000000000017', 'organizations:write', 'organizations', 'write', 'Manage organizations')
ON CONFLICT (name) DO NOTHING;

-- Assign all permissions to ADMIN role
INSERT INTO public.role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000001', id FROM public.permissions
ON CONFLICT DO NOTHING;

-- Assign read permissions to USER role
INSERT INTO public.role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000002', id FROM public.permissions 
WHERE action = 'read' OR name IN ('workspaces:write', 'pages:write', 'alerts:write')
ON CONFLICT DO NOTHING;

-- Assign only read permissions to VIEWER role
INSERT INTO public.role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000003', id FROM public.permissions 
WHERE action = 'read'
ON CONFLICT DO NOTHING;
