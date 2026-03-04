# Auth Module

User authentication, JWT token management, and OAuth providers.

## Domain Entities

- `User` — user profile with status, email verification, notification preferences
- `RefreshToken` — JWT refresh token storage
- `Session` — user session state
- `Role` — role definitions (OWNER, ADMIN, MEMBER, SUPER_ADMIN)

## Use Cases

- `register` — user registration (creates pending registration request)
- `check_subdomain` — validate subdomain availability
- `login` — authenticate with email/password
- `refresh_token` — refresh JWT tokens
- `get_current_user` — fetch authenticated user
- `forgot_password` / `reset_password` — password reset flow

## HTTP Routes (`/auth/*`)

- POST `/auth/register`
- POST `/auth/check-subdomain`
- POST `/auth/login`
- POST `/auth/logout`
- POST `/auth/refresh`
- POST `/auth/forgot-password`
- POST `/auth/reset-password`
- GET `/auth/oauth/{provider}`
- GET `/auth/oauth/{provider}/callback`
- GET `/auth/me` (authenticated)
- PUT `/auth/me`
- PUT `/auth/me/password`
- DELETE `/auth/me`

## Domain Services

- `AuthService` — password hashing/validation (bcrypt)
- `TokenService` — JWT generation and validation

## Infrastructure

- PostgreSQL: `users`, `refresh_tokens`, `sessions`, `roles`, `permissions` tables (public schema)
- OAuth: Google and GitHub providers (conditional on env vars)
- Cookie management with domain/secure flags
- Email: password reset emails
- Event publishing: `user.deleted` event on account deletion
