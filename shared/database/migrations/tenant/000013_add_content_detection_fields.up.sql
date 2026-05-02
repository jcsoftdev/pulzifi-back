ALTER TABLE checks ADD COLUMN IF NOT EXISTS content_block_hash VARCHAR(64) DEFAULT '';
ALTER TABLE monitoring_configs ADD COLUMN IF NOT EXISTS ignore_selectors JSONB NOT NULL DEFAULT '[]'::jsonb;
