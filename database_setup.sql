-- ============================================================
-- PULZIFI DATABASE SETUP
-- Complete database initialization with public schema + tenant template
-- ============================================================

-- ============================================================
-- PHASE 1: PUBLIC SCHEMA TABLES (Shared across all tenants)
-- ============================================================

-- Drop existing objects if they exist (for clean slate)
DROP TRIGGER IF EXISTS after_organization_insert ON public.organizations CASCADE;
DROP FUNCTION IF EXISTS trigger_create_tenant_schema() CASCADE;
DROP FUNCTION IF EXISTS create_tenant_schema(TEXT) CASCADE;
DROP TABLE IF EXISTS public.password_resets CASCADE;
DROP TABLE IF EXISTS public.refresh_tokens CASCADE;
DROP TABLE IF EXISTS public.organization_members CASCADE;
DROP TABLE IF EXISTS public.organizations CASCADE;
DROP TABLE IF EXISTS public.users CASCADE;

-- ============================================================
-- TABLE: users
-- ============================================================
CREATE TABLE public.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    avatar_url TEXT,
    email_verified BOOLEAN DEFAULT FALSE,
    email_notifications_enabled BOOLEAN DEFAULT TRUE,
    notification_frequency VARCHAR(50) DEFAULT 'immediate', -- 'immediate', 'daily_digest', 'weekly_digest'
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_users_email ON public.users(email);
CREATE INDEX idx_users_email_verified ON public.users(email_verified);
CREATE INDEX idx_users_email_notifications_enabled ON public.users(email_notifications_enabled);

-- ============================================================
-- TABLE: organizations
-- ============================================================
CREATE TABLE public.organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(100) UNIQUE NOT NULL,
    schema_name VARCHAR(100) UNIQUE NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_organizations_subdomain ON public.organizations(subdomain);
CREATE INDEX idx_organizations_schema_name ON public.organizations(schema_name);
CREATE INDEX idx_organizations_owner_user_id ON public.organizations(owner_user_id);

-- ============================================================
-- TABLE: organization_members
-- ============================================================
CREATE TABLE public.organization_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES public.organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'MEMBER',
    invited_by UUID REFERENCES public.users(id),
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, user_id)
);

CREATE INDEX idx_organization_members_org_id ON public.organization_members(organization_id);
CREATE INDEX idx_organization_members_user_id ON public.organization_members(user_id);
CREATE INDEX idx_organization_members_user_org ON public.organization_members(user_id, organization_id);

-- ============================================================
-- TABLE: refresh_tokens
-- ============================================================
CREATE TABLE public.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    token VARCHAR(500) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP NULL
);

CREATE INDEX idx_refresh_tokens_user_id ON public.refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON public.refresh_tokens(token);

-- ============================================================
-- TABLE: password_resets
-- ============================================================
CREATE TABLE public.password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    used_at TIMESTAMP NULL
);

CREATE INDEX idx_password_resets_token ON public.password_resets(token);

-- ============================================================
-- PHASE 2: TENANT SCHEMA TEMPLATE FUNCTION
-- ============================================================
-- This function creates the complete tenant schema structure
-- It will be called automatically when a new organization is created

CREATE OR REPLACE FUNCTION create_tenant_schema(schema_name TEXT)
RETURNS VOID AS $$
BEGIN
    -- Create schema
    EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', schema_name);
    
    RAISE NOTICE 'Creating tenant schema: %', schema_name;
    
    -- ========================================
    -- TABLE 1: workspaces
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.workspaces (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name VARCHAR(255) NOT NULL,
            type VARCHAR(50) NOT NULL,
            description TEXT,
            created_by UUID NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
            deleted_at TIMESTAMP NULL
        )', schema_name);
    
    EXECUTE format('CREATE INDEX idx_workspaces_type ON %I.workspaces(type)', schema_name);
    EXECUTE format('CREATE INDEX idx_workspaces_created_by ON %I.workspaces(created_by)', schema_name);
    EXECUTE format('CREATE INDEX idx_workspaces_deleted_at ON %I.workspaces(deleted_at)', schema_name);
    EXECUTE format('CREATE INDEX idx_workspaces_active ON %I.workspaces(deleted_at) WHERE deleted_at IS NULL', schema_name);
    
    -- ========================================
    -- TABLE 2: pages
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.pages (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            workspace_id UUID NOT NULL,
            name VARCHAR(255) NOT NULL,
            url TEXT NOT NULL,
            thumbnail_url TEXT,
            last_checked_at TIMESTAMP NULL,
            last_change_detected_at TIMESTAMP NULL,
            check_count INT DEFAULT 0,
            created_by UUID NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
            deleted_at TIMESTAMP NULL,
            CONSTRAINT fk_workspace FOREIGN KEY (workspace_id) 
                REFERENCES %I.workspaces(id) ON DELETE CASCADE
        )', schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_pages_workspace_id ON %I.pages(workspace_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_pages_url ON %I.pages(url)', schema_name);
    EXECUTE format('CREATE INDEX idx_pages_last_checked_at ON %I.pages(last_checked_at)', schema_name);
    EXECUTE format('CREATE INDEX idx_pages_active ON %I.pages(deleted_at) WHERE deleted_at IS NULL', schema_name);
    
    -- ========================================
    -- TABLE 3: page_tags
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.page_tags (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            page_id UUID NOT NULL,
            tag VARCHAR(100) NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE,
            UNIQUE(page_id, tag)
        )', schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_page_tags_page_id ON %I.page_tags(page_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_page_tags_tag ON %I.page_tags(tag)', schema_name);
    
    -- ========================================
    -- TABLE 4: monitoring_configs
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.monitoring_configs (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            page_id UUID UNIQUE NOT NULL,
            check_frequency VARCHAR(50) NOT NULL DEFAULT ''Every day'',
            schedule_type VARCHAR(50) DEFAULT ''all_time'',
            timezone VARCHAR(100) DEFAULT ''America/Boise'',
            block_ads_cookies BOOLEAN DEFAULT TRUE,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE
        )', schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_monitoring_configs_page_id ON %I.monitoring_configs(page_id)', schema_name);
    
    -- ========================================
    -- TABLE 5: checks
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.checks (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            page_id UUID NOT NULL,
            status VARCHAR(50) NOT NULL,
            screenshot_url TEXT,
            html_snapshot_url TEXT,
            change_detected BOOLEAN DEFAULT FALSE,
            change_type VARCHAR(50),
            error_message TEXT,
            duration_ms INT,
            checked_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE
        )', schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_checks_page_id ON %I.checks(page_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_checks_checked_at ON %I.checks(checked_at)', schema_name);
    EXECUTE format('CREATE INDEX idx_checks_change_detected ON %I.checks(change_detected)', schema_name);
    
    -- ========================================
    -- TABLE 6: alerts
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.alerts (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            workspace_id UUID NOT NULL,
            page_id UUID NOT NULL,
            check_id UUID NOT NULL,
            type VARCHAR(50) NOT NULL,
            title VARCHAR(255) NOT NULL,
            description TEXT,
            read_at TIMESTAMP NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_workspace FOREIGN KEY (workspace_id) 
                REFERENCES %I.workspaces(id) ON DELETE CASCADE,
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE,
            CONSTRAINT fk_check FOREIGN KEY (check_id) 
                REFERENCES %I.checks(id) ON DELETE CASCADE
        )', schema_name, schema_name, schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_alerts_workspace_id ON %I.alerts(workspace_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_alerts_page_id ON %I.alerts(page_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_alerts_type ON %I.alerts(type)', schema_name);
    EXECUTE format('CREATE INDEX idx_alerts_read_at ON %I.alerts(read_at)', schema_name);
    EXECUTE format('CREATE INDEX idx_alerts_created_at ON %I.alerts(created_at)', schema_name);
    
    -- ========================================
    -- TABLE 7: notification_preferences
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.notification_preferences (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            user_id UUID NOT NULL,
            workspace_id UUID REFERENCES %I.workspaces(id) ON DELETE CASCADE,
            page_id UUID REFERENCES %I.pages(id) ON DELETE CASCADE,
            email_enabled BOOLEAN DEFAULT TRUE,
            change_types VARCHAR(100)[],
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT check_workspace_or_page CHECK (
                (workspace_id IS NOT NULL AND page_id IS NULL) OR
                (workspace_id IS NULL AND page_id IS NOT NULL)
            ),
            UNIQUE(user_id, workspace_id),
            UNIQUE(user_id, page_id)
        )', schema_name, schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_notification_preferences_user_id ON %I.notification_preferences(user_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_notification_preferences_workspace_id ON %I.notification_preferences(workspace_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_notification_preferences_page_id ON %I.notification_preferences(page_id)', schema_name);
    
    -- ========================================
    -- TABLE 8: email_logs
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.email_logs (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            alert_id UUID REFERENCES %I.alerts(id) ON DELETE SET NULL,
            recipient_user_id UUID NOT NULL,
            recipient_email VARCHAR(255) NOT NULL,
            subject VARCHAR(500) NOT NULL,
            status VARCHAR(50) NOT NULL,
            provider VARCHAR(50),
            provider_message_id VARCHAR(255),
            error_message TEXT,
            sent_at TIMESTAMP,
            created_at TIMESTAMP NOT NULL DEFAULT NOW()
        )', schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_email_logs_alert_id ON %I.email_logs(alert_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_email_logs_recipient_user_id ON %I.email_logs(recipient_user_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_email_logs_recipient_email ON %I.email_logs(recipient_email)', schema_name);
    EXECUTE format('CREATE INDEX idx_email_logs_status ON %I.email_logs(status)', schema_name);
    EXECUTE format('CREATE INDEX idx_email_logs_sent_at ON %I.email_logs(sent_at)', schema_name);
    EXECUTE format('CREATE INDEX idx_email_logs_created_at ON %I.email_logs(created_at)', schema_name);
    
    -- ========================================
    -- TABLE 9: insights
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.insights (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            page_id UUID NOT NULL,
            check_id UUID NOT NULL,
            insight_type VARCHAR(100) NOT NULL,
            title VARCHAR(255) NOT NULL,
            content TEXT NOT NULL,
            metadata JSONB,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE,
            CONSTRAINT fk_check FOREIGN KEY (check_id) 
                REFERENCES %I.checks(id) ON DELETE CASCADE
        )', schema_name, schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_insights_page_id ON %I.insights(page_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_insights_check_id ON %I.insights(check_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_insights_insight_type ON %I.insights(insight_type)', schema_name);
    EXECUTE format('CREATE INDEX idx_insights_created_at ON %I.insights(created_at)', schema_name);
    
    -- ========================================
    -- TABLE 10: insight_rules
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.insight_rules (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            page_id UUID NOT NULL,
            rule_type VARCHAR(100) NOT NULL,
            enabled BOOLEAN DEFAULT TRUE,
            trigger_condition JSONB,
            created_by UUID NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE
        )', schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_insight_rules_page_id ON %I.insight_rules(page_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_insight_rules_rule_type ON %I.insight_rules(rule_type)', schema_name);
    EXECUTE format('CREATE INDEX idx_insight_rules_enabled ON %I.insight_rules(enabled) WHERE enabled = TRUE', schema_name);
    
    -- ========================================
    -- TABLE 11: reports
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.reports (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            page_id UUID NOT NULL,
            title VARCHAR(255) NOT NULL,
            report_date DATE NOT NULL,
            content JSONB NOT NULL,
            pdf_url TEXT,
            created_by UUID NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE
        )', schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_reports_page_id ON %I.reports(page_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_reports_report_date ON %I.reports(report_date)', schema_name);
    EXECUTE format('CREATE INDEX idx_reports_created_at ON %I.reports(created_at)', schema_name);
    
    -- ========================================
    -- TABLE 12: integrations
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.integrations (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            service_type VARCHAR(50) NOT NULL,
            config JSONB NOT NULL,
            enabled BOOLEAN DEFAULT TRUE,
            created_by UUID NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW()
        )', schema_name);
    
    EXECUTE format('CREATE INDEX idx_integrations_service_type ON %I.integrations(service_type)', schema_name);
    EXECUTE format('CREATE INDEX idx_integrations_enabled ON %I.integrations(enabled) WHERE enabled = TRUE', schema_name);
    
    -- ========================================
    -- TABLE 13: usage_tracking
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.usage_tracking (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            period_start DATE NOT NULL,
            period_end DATE NOT NULL,
            checks_allowed INT NOT NULL,
            checks_used INT DEFAULT 0,
            last_refill_at TIMESTAMP,
            next_refill_at TIMESTAMP,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW()
        )', schema_name);
    
    EXECUTE format('CREATE INDEX idx_usage_tracking_period ON %I.usage_tracking(period_start, period_end)', schema_name);
    
    -- ========================================
    -- TABLE 14: usage_logs
    -- ========================================
    EXECUTE format('
        CREATE TABLE %I.usage_logs (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            page_id UUID NOT NULL,
            check_id UUID NOT NULL,
            checks_consumed INT DEFAULT 1,
            logged_at TIMESTAMP NOT NULL DEFAULT NOW(),
            CONSTRAINT fk_page FOREIGN KEY (page_id) 
                REFERENCES %I.pages(id) ON DELETE CASCADE,
            CONSTRAINT fk_check FOREIGN KEY (check_id) 
                REFERENCES %I.checks(id) ON DELETE CASCADE
        )', schema_name, schema_name, schema_name);
    
    EXECUTE format('CREATE INDEX idx_usage_logs_page_id ON %I.usage_logs(page_id)', schema_name);
    EXECUTE format('CREATE INDEX idx_usage_logs_logged_at ON %I.usage_logs(logged_at)', schema_name);
    
    RAISE NOTICE 'Tenant schema % created successfully with all tables', schema_name;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- PHASE 3: AUTOMATIC TRIGGER FOR TENANT CREATION
-- ============================================================

CREATE OR REPLACE FUNCTION trigger_create_tenant_schema()
RETURNS TRIGGER AS $$
BEGIN
    -- Call function to create tenant schema with full structure
    PERFORM create_tenant_schema(NEW.schema_name);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER after_organization_insert
AFTER INSERT ON public.organizations
FOR EACH ROW
EXECUTE FUNCTION trigger_create_tenant_schema();

-- ============================================================
-- PHASE 4: SAMPLE DATA (Optional - for testing)
-- ============================================================

-- Insert sample user
INSERT INTO public.users (email, password_hash, first_name, last_name, email_verified)
VALUES (
    'admin@pulzifi.com',
    'bcrypt_hash_placeholder',
    'Admin',
    'User',
    TRUE
) ON CONFLICT (email) DO NOTHING;

-- Insert sample organization (this will automatically trigger tenant schema creation)
INSERT INTO public.organizations (name, subdomain, schema_name, owner_user_id)
SELECT
    'jcsoftdev INC',
    'jcsoftdev-inc',
    'jcsoftdev_inc',
    id
FROM public.users
WHERE email = 'admin@pulzifi.com'
ON CONFLICT (subdomain) DO NOTHING;

-- Add user to organization
INSERT INTO public.organization_members (organization_id, user_id, role)
SELECT o.id, u.id, 'ADMIN'
FROM public.organizations o
JOIN public.users u ON u.email = 'admin@pulzifi.com'
WHERE o.subdomain = 'jcsoftdev-inc'
ON CONFLICT (organization_id, user_id) DO NOTHING;

-- ============================================================
-- PHASE 5: VERIFICATION QUERIES
-- ============================================================

-- Verify public schema tables
SELECT 'Public Schema Tables:' as info;
SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename;

-- Verify tenant schema creation
SELECT 'Tenant Schemas Created:' as info;
SELECT schema_name FROM information_schema.schemata 
WHERE schema_name NOT IN ('public', 'information_schema', 'pg_catalog', 'pg_toast')
ORDER BY schema_name;

-- Verify tenant tables (if jcsoftdev_inc schema was created)
SELECT 'Tables in jcsoftdev_inc Schema:' as info;
SELECT tablename FROM pg_tables WHERE schemaname = 'jcsoftdev_inc' ORDER BY tablename;

COMMIT;
