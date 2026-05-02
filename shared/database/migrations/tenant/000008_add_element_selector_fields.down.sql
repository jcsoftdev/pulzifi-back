ALTER TABLE monitoring_configs
  DROP COLUMN IF EXISTS selector_type,
  DROP COLUMN IF EXISTS css_selector,
  DROP COLUMN IF EXISTS xpath_selector,
  DROP COLUMN IF EXISTS selector_offsets;

ALTER TABLE checks
  DROP COLUMN IF EXISTS screenshot_hash,
  DROP COLUMN IF EXISTS vision_change_summary;

ALTER TABLE alerts
  DROP COLUMN IF EXISTS change_summary;
