-- Add element selector fields for region-based monitoring
ALTER TABLE monitoring_configs
  ADD COLUMN IF NOT EXISTS selector_type VARCHAR(20) DEFAULT 'full_page',
  ADD COLUMN IF NOT EXISTS css_selector TEXT DEFAULT '',
  ADD COLUMN IF NOT EXISTS xpath_selector TEXT DEFAULT '',
  ADD COLUMN IF NOT EXISTS selector_offsets JSONB DEFAULT '{"top":0,"right":0,"bottom":0,"left":0}'::jsonb;

-- Add screenshot hash and vision change summary to checks
ALTER TABLE checks
  ADD COLUMN IF NOT EXISTS screenshot_hash VARCHAR(64) DEFAULT '',
  ADD COLUMN IF NOT EXISTS vision_change_summary TEXT DEFAULT '';

-- Add change summary to alerts
ALTER TABLE alerts
  ADD COLUMN IF NOT EXISTS change_summary TEXT DEFAULT '';
