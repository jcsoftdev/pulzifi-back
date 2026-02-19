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
- [ ] Create the PostgreSQL schema: `CREATE SCHEMA IF NOT EXISTS "<schema_name>"` (handled by PG trigger, but verify tenant migrations run)
- [ ] Run tenant migrations against the new schema (reuse logic from `cmd/migrate/main.go:runMigration`)
- [x] Add user as owner in `public.organization_members`
- [x] Assign the `ADMIN` role in `public.user_roles`
- [ ] Assign a default plan in `public.organization_plans`
- [ ] Seed `usage_tracking` in the new tenant schema (currently done by migration `tenant/000004`)
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

- [ ] Fix health endpoint — `string(rune(time.Now().Unix()))` produces garbage Unicode instead of timestamp
- [ ] Fix `SnapshotWorker.createAlert()` — uses raw SQL instead of Alert repository
- [ ] Fix gRPC server — starts but has zero services registered

## 3. Stub Endpoints (returning hardcoded JSON)

### Organization
- [ ] `GET /organizations` — list
- [ ] `GET /organizations/{id}` — get
- [ ] `PUT /organizations/{id}` — update
- [ ] `DELETE /organizations/{id}` — delete
- [ ] gRPC `GetOrganizationBySubdomain` — returns `nil, nil`

### Monitoring
- [ ] `GET /monitoring/checks` — returns empty array
- [ ] `GET /monitoring/checks/{id}` — returns hardcoded mock
- [ ] `GET /monitoring/notification-preferences/{id}` — returns hardcoded mock

### Insight
- [ ] `GET /insights/{id}` — returns hardcoded mock

### Alert
- [ ] `GET /alerts/{id}` — returns hardcoded mock
- [ ] `PUT /alerts/{id}` — returns hardcoded mock
- [ ] `DELETE /alerts/{id}` — returns hardcoded mock

### Integration (entire module is placeholder)
- [ ] `POST /integrations/webhooks`
- [ ] `GET /integrations/webhooks`
- [ ] `GET /integrations/webhooks/{id}`

### Report (entire module is placeholder)
- [ ] `POST /reports`
- [ ] `GET /reports`
- [ ] `GET /reports/{id}`

### Usage
- [ ] `GET /usage/metrics` — returns empty object

## 4. Missing Features

### Team Invite Flow
Currently the invite handler (`modules/team/application/invite_member/handler.go`) returns `ErrUserNotFound` if the invited email doesn't exist. It also sends no email notification.

- [ ] If the invited email doesn't exist, auto-create the user with `status: approved` (no org needed — they're being invited into an existing org)
- [ ] Generate a temporary password or invite token for the new user
- [ ] Send invitation email to the invited user with login/setup link
- [ ] Send email notification to existing users when invited to an organization
- [ ] Frontend: invitation acceptance page (set password if new user)

### Auth
- [ ] Password reset flow (DB table `password_resets` exists, no handler)
- [ ] OAuth/SSO providers

### Email
- [ ] Register email module in `cmd/server/modules.go`
- [ ] Implement real email provider (SendGrid/SES) — currently in-memory only
- [ ] Wire email notifications to relevant events (invites, approvals, alerts)

### Workspace
- [ ] Add endpoint to update workspace member role (frontend has edit dialog, backend missing)

### Organization
- [ ] Implement cascade delete on user deletion (TODO in Kafka subscriber)

### Infrastructure
- [ ] Rate limiting (env vars in `.env.example`, not implemented)
- [ ] Sentry/Datadog integration
- [ ] Slack integration

## 5. Tests

- [ ] Add unit tests for domain services
- [ ] Add unit tests for use case handlers
- [ ] Add unit tests for middleware (auth, tenant, organization)
- [ ] Add repository mocks/interfaces for testing
- [ ] Add table-driven tests
- [ ] Fix existing e2e tests (hardcoded ports, fake JWT)

## 6. Performance

- [ ] Scheduler: batch query across tenant schemas instead of N+1 queries
- [ ] `getPreviousSuccessfulCheck`: add DB query with `LIMIT 1` instead of loading all checks
- [ ] `InsightBroker`: add cache eviction loop (currently grows unbounded)
- [ ] Worker pool: handle full queue gracefully instead of silently dropping jobs
- [ ] `fetchTextFromURL` in snapshot: handle private MinIO URLs, add retry logic

## 7. Database

- [ ] Add `.down.sql` migration files (only 1 of 14 has rollback)
- [ ] Normalize check frequency format (`'1h'` vs `'1 hr'` duplicates in scheduler)

## 8. Security

- [ ] Remove default `JWT_SECRET=secret` — require it to be set
- [ ] Sanitize schema name in `GetSetSearchPathSQL` (string interpolation)

## 9. Frontend Gaps

- [ ] Settings page — currently an empty shell
- [ ] Reports UI — no frontend for reports
- [ ] Integrations UI — no frontend for integrations
- [ ] Hooks directory is empty (`.gitkeep` only)

## 10. Unused Code

- [ ] `shared/kafka/client.go` — Kafka client exists but app uses in-memory EventBus
- [ ] `shared/middleware/health.go` — unused, health defined inline in `main.go`
