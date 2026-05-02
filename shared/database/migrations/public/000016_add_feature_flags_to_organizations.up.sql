ALTER TABLE organizations
  ADD COLUMN IF NOT EXISTS feature_flags JSONB NOT NULL DEFAULT '{}'::jsonb;

UPDATE organizations
   SET feature_flags = feature_flags
       || '{"integrations":{"discord":true,"twilio":true}}'::jsonb
 WHERE deleted_at IS NULL;

ALTER TABLE organizations
  ALTER COLUMN feature_flags
  SET DEFAULT '{"integrations":{"discord":true,"twilio":true}}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_orgs_feature_flags
    ON organizations USING GIN (feature_flags);
