ALTER TABLE checks ADD COLUMN parent_check_id UUID REFERENCES checks(id) ON DELETE CASCADE;
CREATE INDEX idx_checks_parent_check_id ON checks(parent_check_id) WHERE parent_check_id IS NOT NULL;
