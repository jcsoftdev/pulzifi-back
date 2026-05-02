-- Create monitored_sections table for multi-section monitoring
CREATE TABLE IF NOT EXISTS monitored_sections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id         UUID NOT NULL,
    name            TEXT NOT NULL DEFAULT '',
    css_selector    TEXT NOT NULL,
    xpath_selector  TEXT NOT NULL DEFAULT '',
    selector_offsets JSONB DEFAULT '{"top":0,"right":0,"bottom":0,"left":0}'::jsonb,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_monitored_sections_page FOREIGN KEY (page_id)
        REFERENCES pages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_monitored_sections_page_id ON monitored_sections(page_id);

-- Add section_id to checks table to link each check to a specific section
ALTER TABLE checks ADD COLUMN IF NOT EXISTS section_id UUID REFERENCES monitored_sections(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_checks_section_id ON checks(section_id);
