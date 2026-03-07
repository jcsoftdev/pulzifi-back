# Auth Module

User authentication, JWT token management, and OAuth providers.

## Domain Entities

- `User` — user profile with status, email verification, notification preferences
- `RefreshToken` — JWT refresh token storage
- `Session` — user session state
- `Role` — role definitions (OWNER, ADMIN, MEMBER, SUPER_ADMIN)

## Use Cases (application/ directories)

- `register` — user registration (creates pending registration request)
- `check_subdomain` — validate subdomain availability
- `login` — authenticate with email/password
- `refresh_token` — refresh JWT tokens
- `get_current_user` — fetch authenticated user

## HTTP Routes (`/auth/*`)

- POST `/auth/register`
- POST `/auth/check-subdomain`
- POST `/auth/login`
- POST `/auth/logout`
- POST `/auth/refresh`
- POST `/auth/forgot-password` (inline handler, no use case dir)
- POST `/auth/reset-password` (inline handler, no use case dir)
- GET `/auth/oauth/{provider}` — redirect to OAuth provider
- GET `/auth/oauth/{provider}/callback` — OAuth callback
- GET `/auth/me` (authenticated) — get current user
- PUT `/auth/me` (authenticated, inline) — update profile
- PUT `/auth/me/password` (authenticated, inline) — change password
- DELETE `/auth/me` (authenticated, inline) — delete account

## Domain Services

- `AuthService` — password hashing/validation (bcrypt)
- `TokenService` — JWT generation and validation

## Infrastructure

- PostgreSQL: `users`, `refresh_tokens`, `sessions`, `roles`, `permissions` tables (public schema)
- OAuth: Google and GitHub providers (conditional on env vars)
- Cookie management with domain/secure flags
- Email: password reset and notification emails
- Event publishing: `user.deleted` event on account deletion

## Notes

- `forgot_password`, `reset_password`, `update_current_user`, `change_password`, `delete_current_user` routes exist but are implemented inline in module.go (no dedicated use case directories)

## Cross-Module Dependencies (violations)

- Imports `modules/admin/domain/repositories` (RegistrationRequestRepository)
- Imports `modules/email/domain/services` (EmailProvider)
- Imports `modules/email/infrastructure/templates`
- Imports `modules/organization/domain/repositories` (OrganizationRepository)
- Imports `modules/organization/domain/services` (OrganizationService)

**Recommended:** Define email sending and org creation interfaces in this module's domain. Inject implementations from `cmd/server/modules.go`.

## Architecture Improvements

- **Inline handlers should be extracted.** `forgot_password`, `reset_password`, `update_current_user`, `change_password`, `delete_current_user` should each become a use case directory with handler, request, and response files.
- **module.go is 804 lines** — too large. Splitting into use cases will improve maintainability.
- **JWT tokens are stateless** but refresh tokens are in PostgreSQL. For horizontal scaling, consider Redis-backed refresh token storage for faster lookups.
- **Password reset tokens** should have explicit expiry tracking (currently relies on JWT expiry).
