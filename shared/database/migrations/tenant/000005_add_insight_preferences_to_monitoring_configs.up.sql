ALTER TABLE monitoring_configs
  ADD COLUMN IF NOT EXISTS enabled_insight_types    JSONB NOT NULL DEFAULT '["marketing","market_analysis"]'::jsonb,
  ADD COLUMN IF NOT EXISTS enabled_alert_conditions JSONB NOT NULL DEFAULT '["any_changes"]'::jsonb,
  ADD COLUMN IF NOT EXISTS custom_alert_condition   TEXT  NOT NULL DEFAULT '';
