# Auth Module

User authentication, JWT token management, and OAuth providers.

## Domain Entities

- `User` — user profile with status, email verification, notification preferences
- `RefreshToken` — JWT refresh token storage
- `Session` — user session state
- `Role` — role definitions (OWNER, ADMIN, MEMBER, SUPER_ADMIN)

## Use Cases (application/ directories)

11 use case directories exist:

| Directory | Package | Description |
|-----------|---------|-------------|
| `check_subdomain/` | `checksubdomain` | Validate subdomain availability |
| `getcurrentuser/` | `getcurrentuser` | Fetch authenticated user profile |
| `login/` | `login` | Authenticate with email/password |
| `logout/` | `logout` | Revoke refresh token |
| `refreshtoken/` | `refreshtoken` | Refresh JWT tokens (token rotation) |
| `register/` | `register` | User registration (creates pending request) |
| `update_current_user/` | `updatecurrentuser` | Response struct for profile update |
| `forgot_password/` | _(inline)_ | Not extracted yet |
| `reset_password/` | _(inline)_ | Not extracted yet |
| `change_password/` | _(inline)_ | Not extracted yet |
| `delete_current_user/` | _(inline)_ | Not extracted yet |

**Note on package naming:** Directory names use underscores but package names use concatenated lowercase (e.g., `getcurrentuser`, `refreshtoken`, `updatecurrentuser`). This follows the project convention: `directory create_check` → `package createcheck`.

## HTTP Routes (`/auth/*`)

| Method | Path | Handler |
|--------|------|---------|
| POST | `/auth/register` | `register.Handler` |
| POST | `/auth/check-subdomain` | `checksubdomain.Handler` |
| POST | `/auth/login` | `login.Handler` |
| POST | `/auth/logout` | `logout.Handler` |
| POST | `/auth/refresh` | `refreshtoken.Handler` |
| POST | `/auth/forgot-password` | inline in `module.go` |
| POST | `/auth/reset-password` | inline in `module.go` |
| GET | `/auth/oauth/{provider}` | inline OAuth redirect |
| GET | `/auth/oauth/{provider}/callback` | inline OAuth callback |
| GET | `/auth/me` | `getcurrentuser.Handler` |
| PUT | `/auth/me` | inline + `updatecurrentuser.Response` |
| PUT | `/auth/me/password` | inline in `module.go` |
| DELETE | `/auth/me` | inline in `module.go` |

OAuth handlers (`handleOAuthRedirect`, `handleOAuthCallback`) stay inline in `infrastructure/http/module.go` (~430 LOC).

## Domain Services

- `AuthService` — password hashing/validation (bcrypt)
- `TokenService` — JWT generation and validation
- `OrgContextLookup` — interface for combining org identity + plan code + feature flags (for `/auth/me`); implemented in `cmd/wiring/integration/`

## Infrastructure

- PostgreSQL: `users`, `refresh_tokens`, `sessions`, `roles`, `permissions`, `password_resets` tables (public schema)
- OAuth: Google and GitHub providers (conditional on env vars)
- Cookie management with domain/secure flags
- Email: password reset and notification emails via `services.RegistrationNotifier`
- BFF handler: `infrastructure/bff/handler.go` — cross-subdomain cookie/nonce management
- Event publishing: `user.deleted` event on account deletion

## Domain Tests

Tests in `domain/services/` use local fakes (no infrastructure imports):
- `stub_repos_test.go` — in-memory stubs for UserRepository, PermissionRepository, RoleRepository
- `auth_service_test.go` — AuthService interface tests with stubAuthService
- `token_service_test.go` — TokenService interface tests with stubTokenService

## Cross-Module Dependencies (violations)

- Imports `modules/admin/domain/repositories` (RegistrationRequestRepository)
- Imports `modules/email/domain/services` (EmailProvider)
- Imports `modules/email/infrastructure/templates`
- Imports `modules/organization/domain/repositories` (OrganizationRepository)
- Imports `modules/organization/domain/services` (OrganizationService)

**Recommended:** Define email sending and org creation interfaces in this module's domain. Inject implementations from `cmd/server/modules.go`.

## Architecture Improvements

- **Remaining inline handlers should be extracted.** `forgot_password`, `reset_password`, `change_password`, `delete_current_user` should each become a use case directory with handler, request, and response files.
- **`update_current_user/`** has a `response.go` but no `handler.go` yet — the handler is still inline in `module.go`.
- **JWT tokens are stateless** but refresh tokens are in PostgreSQL. For horizontal scaling, consider Redis-backed refresh token storage for faster lookups.
- **Password reset tokens** should have explicit expiry tracking (currently relies on JWT expiry).
