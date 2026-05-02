ALTER TABLE integration_deliveries
  ADD COLUMN IF NOT EXISTS attempt_history JSONB NOT NULL DEFAULT '[]'::jsonb;
