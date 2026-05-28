ALTER TABLE usage_tracking
    DROP COLUMN IF EXISTS ai_insights_used,
    DROP COLUMN IF EXISTS ai_insights_allowed;
