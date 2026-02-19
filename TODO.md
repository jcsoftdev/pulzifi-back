# Pulzifi Backend — Task Tracker

## 1. User Registration & Approval Flow (New)

Currently: registration creates an orphan user (no org, no role, no tenant). Login has no status check. Tenant schemas are only created via the CLI migration tool (`cmd/migrate/main.go`). The `create_organization` handler only inserts into `public.organizations` — it does NOT create the PG schema, add the user to `organization_members`, or run tenant migrations.

### Registration Changes
- [x] Add `organization_name` and `organization_subdomain` fields to registration request (`modules/auth/application/register/request.go`)
- [x] Add `status` field to `User` entity (`modules/auth/domain/entities/user.go`) — values: `pending`, `approved`, `rejected`
- [x] Add DB migration: `ALTER TABLE public.users ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'approved'` (default 'approved' for existing users)
- [x] Store requested org name/subdomain during registration (`public.registration_requests` table via `modules/admin/`)
- [x] Update register handler to save org info alongside user creation (`modules/auth/application/register/handler.go`)
- [x] Block login for users with `status != 'approved'` — return 403 in auth HTTP module (`modules/auth/infrastructure/services/bcrypt_auth_service.go`)

### Super Admin Approval Endpoints
- [x] `GET /admin/users/pending` — list users awaiting approval (super admin only)
- [x] `PUT /admin/users/{id}/approve` — approve a user (super admin only)
- [x] `PUT /admin/users/{id}/reject` — reject a user (super admin only)

### On Approval — Tenant Provisioning
When a super admin approves a user, the system must (in a single transaction):
- [x] Set user `status` to `approved`
- [x] Create the organization in `public.organizations` (name + subdomain from registration) — PG trigger `after_organization_insert` auto-creates tenant schema
- [x] Generate schema name from subdomain (via `OrganizationService.GenerateSchemaName`)
- [x] Create the PostgreSQL schema: `CREATE SCHEMA IF NOT EXISTS "<schema_name>"` (handled by PG trigger, but verify tenant migrations run)
- [x] Run tenant migrations against the new schema (extracted to `shared/database/migrator.go:ProvisionTenantSchema`)
- [x] Add user as owner in `public.organization_members`
- [x] Assign the `ADMIN` role in `public.user_roles`
- [x] Assign a default plan in `public.organization_plans`
- [x] Seed `usage_tracking` in the new tenant schema (handled by tenant migration `000004` which runs via `ProvisionTenantSchema`)
- [ ] (Optional) Send approval notification email to the user

### Frontend (backend API ready, frontend not yet)
- [ ] Update registration form to include `organization_name` and `organization_subdomain` fields (`POST /auth/register`)
- [ ] Show "pending approval" message after registration (API returns `status: "pending"`)
- [ ] Handle 403 on login for pending/rejected users — show appropriate message (API returns `code: "USER_NOT_APPROVED"` or `"USER_REJECTED"`)
- [ ] Block dashboard access for unapproved users (redirect to a "pending approval" page)
- [ ] Super admin panel — new page `/admin/users`:
  - [ ] Table listing pending users via `GET /admin/users/pending` (shows email, name, org name, subdomain, date)
  - [ ] Approve button per row → `PUT /admin/users/{id}/approve`
  - [ ] Reject button per row → `PUT /admin/users/{id}/reject`
  - [ ] Only visible to users with `SUPER_ADMIN` role (hide nav link for others)

---

## 2. Critical Bugs

- [x] Fix health endpoint — `string(rune(time.Now().Unix()))` produces garbage Unicode instead of timestamp
- [x] Fix `SnapshotWorker.createAlert()` — uses raw SQL instead of Alert repository
- [x] Fix gRPC server — register OrganizationServiceServer (CreateOrg, GetOrg, GetOrgBySubdomain)

## 3. Stub Endpoints (returning hardcoded JSON)

### Organization
- [x] `GET /organizations` — list (by current user)
- [x] `GET /organizations/{id}` — get
- [x] `PUT /organizations/{id}` — update name
- [x] `DELETE /organizations/{id}` — soft delete
- [x] gRPC `GetOrganizationBySubdomain` — implemented

### Monitoring
- [x] `GET /monitoring/checks` — list by page_id query param
- [x] `GET /monitoring/checks/{id}` — get by ID
- [x] `GET /monitoring/notification-preferences/{id}` — get by ID

### Insight
- [x] `GET /insights/{id}` — real implementation

### Alert
- [x] `GET /alerts/{id}` — real implementation
- [x] `PUT /alerts/{id}` — marks alert as read
- [x] `DELETE /alerts/{id}` — real implementation

### Integration
- [x] `POST /integrations` — upsert integration
- [x] `GET /integrations` — list integrations
- [x] `DELETE /integrations/{id}` — delete integration
- [x] `POST /integrations/webhooks` — create webhook integration
- [x] `GET /integrations/webhooks` — list webhook integrations
- [x] `GET /integrations/webhooks/{id}` — get webhook by ID

### Report
- [x] `POST /reports` — real implementation (DDD: entity, repository, postgres persistence)
- [x] `GET /reports` — real implementation (supports `?page_id=` filter)
- [x] `GET /reports/{id}` — real implementation

### Usage
- [x] `GET /usage/metrics` — real implementation (checks, pages, workspaces, alerts stats)

## 4. Missing Features

### Team Invite Flow
- [x] If the invited email doesn't exist, auto-create the user with `status: approved` (no org needed — they're being invited into an existing org)
- [x] Generate a temporary password (random, bcrypt-hashed) for the new user
- [ ] Send invitation email to the invited user with login/setup link
- [ ] Send email notification to existing users when invited to an organization
- [ ] Frontend: invitation acceptance page (set password if new user)

### Auth
- [x] Password reset flow — full implementation in `modules/auth/infrastructure/http/module.go`: generates token, stores in `public.password_resets`, sends email, verifies token, updates password
- [x] OAuth/SSO providers — GitHub and Google OAuth2 implemented in `modules/auth/infrastructure/oauth/`, wired at `/auth/oauth/{provider}` and `/auth/oauth/{provider}/callback`

### Email
- [x] Register email module in `cmd/server/modules.go`
- [x] Implement real email provider — Resend (`modules/email/infrastructure/providers/resend_provider.go`), selected when `RESEND_API_KEY` is set; falls back to no-op
- [ ] Wire email notifications to relevant events (approvals, alerts) — invites partially wired, approvals not

### Workspace
- [x] `PUT /workspaces/{id}/members/{user_id}` — update workspace member role (`modules/workspace/application/update_member_role/`)

### Organization
- [ ] Implement cascade delete on user deletion (TODO in messaging subscriber)

### Infrastructure
- [ ] Rate limiting (env vars in `.env.example`, not implemented)
- [ ] Sentry/Datadog integration
- [ ] Slack integration

## 5. Tests

- [x] Unit tests for auth register handler (`modules/auth/application/register/handler_test.go`)
- [x] Unit tests for auth middleware (`modules/auth/infrastructure/middleware/auth_middleware_test.go`)
- [x] Unit tests for workspace update_member_role handler (`modules/workspace/application/update_member_role/handler_test.go`)
- [x] Unit tests for workspace role value object (`modules/workspace/domain/value_objects/workspace_role_test.go`)
- [x] Unit tests for organization domain service (`modules/organization/domain/services/organization_service_test.go`)
- [x] Unit tests for email domain service (`modules/email/domain/services/email_service_test.go`)
- [x] Repository mocks for admin, auth, organization, workspace modules
- [ ] Unit tests for remaining domain services (snapshot, monitoring, insight, alert)
- [ ] Unit tests for remaining use case handlers
- [ ] Unit tests for tenant middleware (`shared/middleware/tenant_test.go` exists but coverage incomplete)
- [ ] Fix existing e2e tests (hardcoded ports, fake JWT)

## 6. Performance

- [ ] Scheduler: batch query across tenant schemas instead of N+1 queries
- [x] `getPreviousSuccessfulCheck`: use `GetPreviousSuccessfulByPage` with `LIMIT 1` instead of loading all checks
- [x] `InsightBroker`: add cache eviction loop (ticker every cacheTTL, removes expired entries)
- [ ] Worker pool: handle full queue gracefully instead of silently dropping jobs
- [ ] `fetchTextFromURL` in snapshot: handle private MinIO URLs, add retry logic

## 7. Database

- [ ] Add `.down.sql` migration files (only 1 of 14 has rollback)
- [ ] Normalize check frequency format (`'1h'` vs `'1 hr'` duplicates in scheduler)

## 8. Security

- [x] Remove default `JWT_SECRET=secret` — fatal in production, warning + insecure default in development
- [x] Sanitize schema name in `GetSetSearchPathSQL` — validate with regex, reject invalid names

## 9. Frontend Gaps

- [ ] Settings page — currently an empty shell
- [ ] Reports UI — no frontend for reports
- [ ] Integrations UI — no frontend for integrations
- [ ] Hooks directory is empty (`.gitkeep` only)

## 10. Unused Code

- [x] `shared/kafka/client.go` — deleted (app uses in-memory EventBus)
- [ ] `shared/middleware/health.go` — unused, health defined inline in `main.go`
