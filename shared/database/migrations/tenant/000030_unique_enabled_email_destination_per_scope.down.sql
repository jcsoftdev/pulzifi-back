-- Rollback: unique_enabled_email_destination_per_scope
-- Scope: tenant

DROP INDEX IF EXISTS uq_integration_destinations_enabled_email_scope;
