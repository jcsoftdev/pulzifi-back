-- Reverse 000001_init_tenant_schema.up.sql
-- Drop tables in reverse dependency order

DROP INDEX IF EXISTS idx_usage_logs_logged_at;
DROP INDEX IF EXISTS idx_usage_logs_page_id;
DROP TABLE IF EXISTS usage_logs;

DROP INDEX IF EXISTS idx_usage_tracking_period;
DROP TABLE IF EXISTS usage_tracking;

DROP INDEX IF EXISTS idx_integrations_service_type;
DROP TABLE IF EXISTS integrations;

DROP INDEX IF EXISTS idx_reports_created_at;
DROP INDEX IF EXISTS idx_reports_report_date;
DROP INDEX IF EXISTS idx_reports_page_id;
DROP TABLE IF EXISTS reports;

DROP INDEX IF EXISTS idx_insight_rules_rule_type;
DROP INDEX IF EXISTS idx_insight_rules_page_id;
DROP TABLE IF EXISTS insight_rules;

DROP INDEX IF EXISTS idx_insights_created_at;
DROP INDEX IF EXISTS idx_insights_insight_type;
DROP INDEX IF EXISTS idx_insights_check_id;
DROP INDEX IF EXISTS idx_insights_page_id;
DROP TABLE IF EXISTS insights;

DROP INDEX IF EXISTS idx_email_logs_created_at;
DROP INDEX IF EXISTS idx_email_logs_sent_at;
DROP INDEX IF EXISTS idx_email_logs_status;
DROP INDEX IF EXISTS idx_email_logs_recipient_email;
DROP INDEX IF EXISTS idx_email_logs_recipient_user_id;
DROP INDEX IF EXISTS idx_email_logs_alert_id;
DROP TABLE IF EXISTS email_logs;

DROP INDEX IF EXISTS idx_notification_preferences_page_id;
DROP INDEX IF EXISTS idx_notification_preferences_workspace_id;
DROP INDEX IF EXISTS idx_notification_preferences_user_id;
DROP TABLE IF EXISTS notification_preferences;

DROP INDEX IF EXISTS idx_alerts_created_at;
DROP INDEX IF EXISTS idx_alerts_read_at;
DROP INDEX IF EXISTS idx_alerts_type;
DROP INDEX IF EXISTS idx_alerts_page_id;
DROP INDEX IF EXISTS idx_alerts_workspace_id;
DROP TABLE IF EXISTS alerts;

DROP INDEX IF EXISTS idx_checks_section_id;
DROP INDEX IF EXISTS idx_checks_change_detected;
DROP INDEX IF EXISTS idx_checks_checked_at;
DROP INDEX IF EXISTS idx_checks_page_id;
DROP TABLE IF EXISTS checks;

DROP INDEX IF EXISTS idx_monitored_sections_page_id;
DROP TABLE IF EXISTS monitored_sections;

DROP INDEX IF EXISTS idx_monitoring_configs_page_id;
DROP TABLE IF EXISTS monitoring_configs;

DROP INDEX IF EXISTS idx_page_tags_tag;
DROP INDEX IF EXISTS idx_page_tags_page_id;
DROP TABLE IF EXISTS page_tags;

DROP INDEX IF EXISTS idx_pages_last_checked_at;
DROP INDEX IF EXISTS idx_pages_url;
DROP INDEX IF EXISTS idx_pages_workspace_id;
DROP TABLE IF EXISTS pages;

DROP INDEX IF EXISTS idx_workspace_members_role;
DROP INDEX IF EXISTS idx_workspace_members_user_id;
DROP TABLE IF EXISTS workspace_members;

DROP INDEX IF EXISTS idx_workspaces_deleted_at;
DROP INDEX IF EXISTS idx_workspaces_created_by;
DROP INDEX IF EXISTS idx_workspaces_type;
DROP TABLE IF EXISTS workspaces;
