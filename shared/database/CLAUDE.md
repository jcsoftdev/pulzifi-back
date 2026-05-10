# Database Package (`shared/database/`)

PostgreSQL connection pool and tenant schema provisioning.

## Files

- `connection.go` — Database connection with retry logic
- `migrator.go` — Tenant schema creation and migration
- `migrations/public/` — 17 public schema migrations
- `migrations/tenant/` — 20 tenant schema migrations

## Exported API

### Functions
- `Connect(cfg *config.Config) (*sql.DB, error)` — Creates PostgreSQL connection pool with exponential backoff retry (2s initial, 5 retries, doubling). Checks `DATABASE_URL` env first, falls back to individual config fields. Sets `MaxOpenConns` from config, `MaxIdleConns` to half.
- `ProvisionTenantSchema(db *sql.DB, schemaName string) error` — Validates schema name (regex `^[a-zA-Z_][a-zA-Z0-9_]*$`), creates schema (`CREATE SCHEMA IF NOT EXISTS`), runs tenant migrations. Idempotent — safe to call multiple times.

## Migrations

### Public Schema (17)
1. `init_public_schema` — Users, organizations base tables
2. `init_roles_permissions` — Roles and permissions tables
3. `seed_roles_permissions` — Seed role/permission data
4. `add_super_admin_plans` — Plan definitions
5. `seed_default_admin_organization` — Default admin org
6. `ensure_default_org_tenant_schema` — Default org tenant
7. `force_default_tenant_structure` — Force default structure
8. `add_sessions_table` — User sessions
9. `add_storage_period_to_plans` — Storage period
10. `add_user_status_and_registration_requests` — User approval workflow
11. `add_oauth_providers` — OAuth provider support
12. `add_invitation_status_to_members` — Invitation status tracking
13. `add_super_admin_invitations` — Super admin invitation support
14. `assign_starter_plan_to_orgs` — Assign starter plan to existing orgs
15. `create_integrations_oauth` — Integration OAuth token storage (public schema)
16. `add_feature_flags_to_organizations` — Per-org `feature_flags` JSONB column
17. `create_integration_send_quotas` — Monthly quota tracking for integration providers

### Tenant Schema (20)
1. `init_tenant_schema` — Workspaces, pages, checks, monitoring base tables
2. `add_tags_to_workspaces` — Workspace tags
3. `add_content_hash_to_checks` — Content hash for change detection
4. `seed_usage_tracking_from_plan` — Usage tracking initialization
5. `add_insight_preferences_to_monitoring_configs` — Insight type preferences
6. `normalize_frequencies` — Monitoring frequency normalization
7. `fix_usage_tracking_billing_period` — Billing period fix
8. `add_element_selector_fields` — CSS/XPath selector fields
9. `add_change_summary_to_alerts` — Alert change summaries
10. `add_monitored_sections` — Page section monitoring
11. `add_rect_to_monitored_sections` — Bounding rectangles for sections
12. `add_parent_check_id` — Parent-child check relationships
13. `add_content_detection_fields` — Content detection metadata
14. `add_content_diff_to_checks` — Text diff between check snapshots
15. `add_diff_image_url_to_checks` — Visual diff image URL
16. `ensure_current_billing_period` — Backfill current billing period rows
17. `create_integration_destinations` — Delivery destination table (tenant-scoped)
18. `create_integration_deliveries` — Delivery history table
19. `drop_legacy_integrations` — Remove old integrations table schema
20. `add_attempt_history_to_deliveries` — Per-attempt history JSON on deliveries

## Scaffold New Migrations

```bash
./tools/scripts/new-migration.sh <scope> <description>
# scope: public or tenant
```

## Dependencies

- `database/sql`, `lib/pq`
- `golang-migrate/migrate/v4`
- `shared/config`, `shared/logger`
