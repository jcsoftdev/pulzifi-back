-- Reverse 000002_add_tags_to_workspaces.up.sql

ALTER TABLE workspaces DROP COLUMN IF EXISTS tags;
