# Admin Module

Manages user registration requests and admin approval workflow.

## Domain Entities

- `RegistrationRequest` — pending registration with user info, org details, and approval status

## Use Cases (application/ directories)

- `list_pending_users` — list pending registrations (with pagination)
- `approve_user` — approve a registration and create organization
- `reject_user` — reject a registration request

## HTTP Routes (`/admin/*`, requires SUPER_ADMIN role)

- GET `/admin/users/pending` — list pending registrations
- PUT `/admin/users/{id}/approve` — approve registration
- PUT `/admin/users/{id}/reject` — reject registration

## Infrastructure

- PostgreSQL: `registration_requests` table (public schema)
- Email: sends approval/rejection notifications via templates
- Cross-module: integrates with Organization, Auth, Email modules

## Cross-Module Dependencies (violations)

- Imports `modules/auth/domain/repositories` (UserRepository)
- Imports `modules/auth/infrastructure/middleware` (AuthMiddleware)
- Imports `modules/email/domain/services` (EmailProvider)
- Imports `modules/email/infrastructure/templates`
- Imports `modules/organization/domain/repositories` (OrganizationRepository)
- Imports `modules/organization/domain/services` (OrganizationService)

**Recommended:** Define interfaces for user management, email sending, and org provisioning in this module's domain. Inject implementations from `cmd/server/modules.go`.
