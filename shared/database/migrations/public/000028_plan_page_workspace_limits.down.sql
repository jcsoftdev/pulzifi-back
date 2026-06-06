ALTER TABLE public.plans
    DROP COLUMN IF EXISTS max_pages,
    DROP COLUMN IF EXISTS max_workspaces;
