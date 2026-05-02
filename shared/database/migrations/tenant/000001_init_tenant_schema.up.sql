-- Create workspaces table
CREATE TABLE IF NOT EXISTS workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    tags TEXT[],
    created_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE INDEX IF NOT EXISTS idx_workspaces_type ON workspaces(type);
CREATE INDEX IF NOT EXISTS idx_workspaces_created_by ON workspaces(created_by);
CREATE INDEX IF NOT EXISTS idx_workspaces_deleted_at ON workspaces(deleted_at);

-- Create workspace_members table
CREATE TABLE IF NOT EXISTS workspace_members (
    workspace_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('owner', 'editor', 'viewer')),
    invited_by UUID,
    invited_at TIMESTAMP NOT NULL DEFAULT NOW(),
    removed_at TIMESTAMP NULL,
    PRIMARY KEY (workspace_id, user_id),
    CONSTRAINT fk_workspace FOREIGN KEY (workspace_id)
        REFERENCES workspaces(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_workspace_members_user_id ON workspace_members(user_id);
CREATE INDEX IF NOT EXISTS idx_workspace_members_role ON workspace_members(workspace_id, role);

-- Create pages table
CREATE TABLE IF NOT EXISTS pages (
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
        REFERENCES workspaces(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_pages_workspace_id ON pages(workspace_id);
CREATE INDEX IF NOT EXISTS idx_pages_url ON pages(url);
CREATE INDEX IF NOT EXISTS idx_pages_last_checked_at ON pages(last_checked_at);

-- Create page_tags table
CREATE TABLE IF NOT EXISTS page_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID NOT NULL,
    tag VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_page FOREIGN KEY (page_id)
        REFERENCES pages(id) ON DELETE CASCADE,
    UNIQUE(page_id, tag)
);

CREATE INDEX IF NOT EXISTS idx_page_tags_page_id ON page_tags(page_id);
CREATE INDEX IF NOT EXISTS idx_page_tags_tag ON page_tags(tag);

-- Create monitoring_configs table
CREATE TABLE IF NOT EXISTS monitoring_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID UNIQUE NOT NULL,
    check_frequency VARCHAR(50) NOT NULL DEFAULT '24h',
    schedule_type VARCHAR(50) DEFAULT 'all_time',
    timezone VARCHAR(100) DEFAULT 'America/Boise',
    block_ads_cookies BOOLEAN DEFAULT TRUE,
    enabled_insight_types JSONB NOT NULL DEFAULT '["marketing","market_analysis"]'::jsonb,
    enabled_alert_conditions JSONB NOT NULL DEFAULT '["any_changes"]'::jsonb,
    custom_alert_condition TEXT NOT NULL DEFAULT '',
    selector_type VARCHAR(20) DEFAULT 'full_page',
    css_selector TEXT DEFAULT '',
    xpath_selector TEXT DEFAULT '',
    selector_offsets JSONB DEFAULT '{"top":0,"right":0,"bottom":0,"left":0}'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_page FOREIGN KEY (page_id)
        REFERENCES pages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_monitoring_configs_page_id ON monitoring_configs(page_id);

-- Create monitored_sections table (must precede checks due to FK reference)
CREATE TABLE IF NOT EXISTS monitored_sections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id         UUID NOT NULL,
    name            TEXT NOT NULL DEFAULT '',
    css_selector    TEXT NOT NULL,
    xpath_selector  TEXT NOT NULL DEFAULT '',
    selector_offsets JSONB DEFAULT '{"top":0,"right":0,"bottom":0,"left":0}'::jsonb,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_monitored_sections_page FOREIGN KEY (page_id)
        REFERENCES pages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_monitored_sections_page_id ON monitored_sections(page_id);

-- Create checks table
CREATE TABLE IF NOT EXISTS checks (
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
    content_hash VARCHAR(64),
    screenshot_hash VARCHAR(64) DEFAULT '',
    vision_change_summary TEXT DEFAULT '',
    section_id UUID REFERENCES monitored_sections(id) ON DELETE SET NULL,
    CONSTRAINT fk_page FOREIGN KEY (page_id)
        REFERENCES pages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_checks_page_id ON checks(page_id);
CREATE INDEX IF NOT EXISTS idx_checks_checked_at ON checks(checked_at);
CREATE INDEX IF NOT EXISTS idx_checks_change_detected ON checks(change_detected);
CREATE INDEX IF NOT EXISTS idx_checks_section_id ON checks(section_id);

-- Create alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    page_id UUID NOT NULL,
    check_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    metadata JSONB,
    change_summary TEXT DEFAULT '',
    read_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_workspace FOREIGN KEY (workspace_id)
        REFERENCES workspaces(id) ON DELETE CASCADE,
    CONSTRAINT fk_page FOREIGN KEY (page_id)
        REFERENCES pages(id) ON DELETE CASCADE,
    CONSTRAINT fk_check FOREIGN KEY (check_id)
        REFERENCES checks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_alerts_workspace_id ON alerts(workspace_id);
CREATE INDEX IF NOT EXISTS idx_alerts_page_id ON alerts(page_id);
CREATE INDEX IF NOT EXISTS idx_alerts_type ON alerts(type);
CREATE INDEX IF NOT EXISTS idx_alerts_read_at ON alerts(read_at);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at);

-- Create notification_preferences table
CREATE TABLE IF NOT EXISTS notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
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
);

CREATE INDEX IF NOT EXISTS idx_notification_preferences_user_id ON notification_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_notification_preferences_workspace_id ON notification_preferences(workspace_id);
CREATE INDEX IF NOT EXISTS idx_notification_preferences_page_id ON notification_preferences(page_id);

-- Create email_logs table
CREATE TABLE IF NOT EXISTS email_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id UUID REFERENCES alerts(id) ON DELETE SET NULL,
    recipient_user_id UUID NOT NULL,
    recipient_email VARCHAR(255) NOT NULL,
    subject VARCHAR(500) NOT NULL,
    status VARCHAR(50) NOT NULL,
    provider VARCHAR(50),
    provider_message_id VARCHAR(255),
    error_message TEXT,
    sent_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_logs_alert_id ON email_logs(alert_id);
CREATE INDEX IF NOT EXISTS idx_email_logs_recipient_user_id ON email_logs(recipient_user_id);
CREATE INDEX IF NOT EXISTS idx_email_logs_recipient_email ON email_logs(recipient_email);
CREATE INDEX IF NOT EXISTS idx_email_logs_status ON email_logs(status);
CREATE INDEX IF NOT EXISTS idx_email_logs_sent_at ON email_logs(sent_at);
CREATE INDEX IF NOT EXISTS idx_email_logs_created_at ON email_logs(created_at);

-- Create insights table
CREATE TABLE IF NOT EXISTS insights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID NOT NULL,
    check_id UUID NOT NULL,
    insight_type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_page FOREIGN KEY (page_id)
        REFERENCES pages(id) ON DELETE CASCADE,
    CONSTRAINT fk_check FOREIGN KEY (check_id)
        REFERENCES checks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_insights_page_id ON insights(page_id);
CREATE INDEX IF NOT EXISTS idx_insights_check_id ON insights(check_id);
CREATE INDEX IF NOT EXISTS idx_insights_insight_type ON insights(insight_type);
CREATE INDEX IF NOT EXISTS idx_insights_created_at ON insights(created_at);

-- Create insight_rules table
CREATE TABLE IF NOT EXISTS insight_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID NOT NULL,
    rule_type VARCHAR(100) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    trigger_condition JSONB,
    created_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_page FOREIGN KEY (page_id)
        REFERENCES pages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_insight_rules_page_id ON insight_rules(page_id);
CREATE INDEX IF NOT EXISTS idx_insight_rules_rule_type ON insight_rules(rule_type);

-- Create reports table
CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    report_date DATE NOT NULL,
    content JSONB NOT NULL,
    pdf_url TEXT,
    created_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    CONSTRAINT fk_page FOREIGN KEY (page_id)
        REFERENCES pages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_reports_page_id ON reports(page_id);
CREATE INDEX IF NOT EXISTS idx_reports_report_date ON reports(report_date);
CREATE INDEX IF NOT EXISTS idx_reports_created_at ON reports(created_at);

-- Create integrations table
CREATE TABLE IF NOT EXISTS integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    created_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE INDEX IF NOT EXISTS idx_integrations_service_type ON integrations(service_type);

-- Create usage_tracking table
CREATE TABLE IF NOT EXISTS usage_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    checks_allowed INT NOT NULL,
    checks_used INT DEFAULT 0,
    last_refill_at TIMESTAMP,
    next_refill_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_usage_tracking_period ON usage_tracking(period_start, period_end);

-- Create usage_logs table
CREATE TABLE IF NOT EXISTS usage_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID NOT NULL,
    check_id UUID NOT NULL,
    checks_consumed INT DEFAULT 1,
    logged_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_page FOREIGN KEY (page_id)
        REFERENCES pages(id) ON DELETE CASCADE,
    CONSTRAINT fk_check FOREIGN KEY (check_id)
        REFERENCES checks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_usage_logs_page_id ON usage_logs(page_id);
CREATE INDEX IF NOT EXISTS idx_usage_logs_logged_at ON usage_logs(logged_at);
