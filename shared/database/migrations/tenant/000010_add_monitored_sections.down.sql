-- Remove section_id from checks
ALTER TABLE checks DROP COLUMN IF EXISTS section_id;

-- Drop monitored_sections table
DROP TABLE IF EXISTS monitored_sections;
