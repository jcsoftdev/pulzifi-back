-- Reverse 000008_add_sessions_table.up.sql

DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP TABLE IF EXISTS public.sessions;
