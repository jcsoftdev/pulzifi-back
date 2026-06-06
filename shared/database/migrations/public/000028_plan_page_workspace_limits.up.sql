-- Per-plan page and workspace caps. NULL = unlimited (Enterprise/trial).
-- Mirrors the ai_insights_allowed_monthly pattern from migration 000022.
-- Idempotent: ADD COLUMN IF NOT EXISTS.
ALTER TABLE public.plans
    ADD COLUMN IF NOT EXISTS max_pages      INTEGER NULL,
    ADD COLUMN IF NOT EXISTS max_workspaces INTEGER NULL;

UPDATE public.plans SET max_pages = 1,    max_workspaces = 1,    updated_at = NOW() WHERE code = 'free';
UPDATE public.plans SET max_pages = 22,   max_workspaces = 1,    updated_at = NOW() WHERE code = 'starter';
UPDATE public.plans SET max_pages = 45,   max_workspaces = 10,   updated_at = NOW() WHERE code = 'pro';
UPDATE public.plans SET max_pages = NULL, max_workspaces = NULL, updated_at = NOW() WHERE code = 'enterprise';
UPDATE public.plans SET max_pages = NULL, max_workspaces = NULL, updated_at = NOW() WHERE code = 'trial';
