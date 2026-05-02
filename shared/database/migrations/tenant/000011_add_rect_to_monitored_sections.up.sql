ALTER TABLE monitored_sections
  ADD COLUMN rect JSONB,
  ADD COLUMN viewport_width INTEGER;
