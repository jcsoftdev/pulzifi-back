-- Force complete tenant structure for default organization when helper function exists

DO $$
BEGIN
    -- If core tenant table is missing, enforce tenant structure creation for default schema
    IF to_regclass('jcsoftdev_inc.workspaces') IS NULL THEN
        IF EXISTS (
            SELECT 1
            FROM pg_proc p
            JOIN pg_namespace n ON n.oid = p.pronamespace
            WHERE n.nspname = 'public'
              AND p.proname = 'create_tenant_schema'
              AND p.pronargs = 1
        ) THEN
            PERFORM public.create_tenant_schema('jcsoftdev_inc');
        ELSE
            CREATE SCHEMA IF NOT EXISTS jcsoftdev_inc;
        END IF;
    END IF;
END;
$$;
