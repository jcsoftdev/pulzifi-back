-- Migration: unique_enabled_email_destination_per_scope
-- Scope: tenant
-- Created: 2026-06-22T15:06:57Z

-- Enforce at most ONE enabled email destination per scope (org/workspace/page).
-- The change-alert dispatcher seeds a default email destination from workspace
-- members when none is configured; under the worker's concurrent processing two
-- sibling pages could otherwise race and seed duplicates, producing duplicate
-- alert emails. This partial unique index makes the seeding Create fail atomically
-- for the race loser instead. It also matches the product model: one email
-- destination (holding a list of recipient emails) per scope.
CREATE UNIQUE INDEX IF NOT EXISTS uq_integration_destinations_enabled_email_scope
    ON integration_destinations (scope_type, scope_id)
    WHERE service_type = 'email' AND enabled;
