-- Reverse 000009_add_storage_period_to_plans.up.sql

ALTER TABLE public.plans DROP COLUMN IF EXISTS storage_period_days;
