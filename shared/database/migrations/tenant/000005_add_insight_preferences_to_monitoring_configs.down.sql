-- Reverse 000005_add_insight_preferences_to_monitoring_configs.up.sql

ALTER TABLE monitoring_configs DROP COLUMN IF EXISTS custom_alert_condition;
ALTER TABLE monitoring_configs DROP COLUMN IF EXISTS enabled_alert_conditions;
ALTER TABLE monitoring_configs DROP COLUMN IF EXISTS enabled_insight_types;
