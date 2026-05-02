DROP TABLE IF EXISTS public.registration_requests;
DROP INDEX IF EXISTS public.idx_users_status;
ALTER TABLE public.users DROP COLUMN IF EXISTS status;
