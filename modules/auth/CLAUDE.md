# Auth Module

## Responsibility

User authentication, JWT token management, OAuth2 integration (Google, GitHub), password reset/recovery, and session management.

## Entities

- **User** — ID, Email, PasswordHash, FirstName, LastName, Status (pending/approved/rejected), EmailVerified, NotificationFrequency
- **Session** — ID, UserID, ExpiresAt
- **RefreshToken** — ID, UserID, Token, ExpiresAt, IsRevoked
- **Role** — ID, Name, Description
- **Permission** — ID, Name, Resource, Action

## Repository Interfaces

- `UserRepository` — CRUD + GetByEmail, ExistsByEmail, UpdateStatus, ListByStatus
- `SessionRepository` — Create, FindByID, DeleteByID, DeleteExpired
- `RefreshTokenRepository` — Create, FindByToken, FindByUserID, Revoke, RevokeAllByUserID, DeleteExpired
- `RoleRepository`, `PermissionRepository`

## Routes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/register` | User registration |
| POST | `/auth/login` | Login, returns JWT tokens |
| POST | `/auth/logout` | Clear session |
| POST | `/auth/refresh` | Refresh JWT |
| POST | `/auth/check-subdomain` | Check subdomain availability |
| POST | `/auth/forgot-password` | Send reset email |
| POST | `/auth/reset-password` | Reset with token |
| GET/POST | `/auth/oauth/{provider}` | OAuth redirect and callback |
| GET | `/auth/me` | Get current user |
| PUT | `/auth/me` | Update profile |
| PUT | `/auth/me/password` | Change password |
| DELETE | `/auth/me` | Delete account |

## Dependencies

- Admin module (RegistrationRequestRepository)
- Organization module (creates org on approval)
- Email module (password reset, verification)
- EventBus (user.deleted event)
- OAuth providers (Google, GitHub)

## Constraints

- Passwords hashed with bcrypt
- JWT tokens are HttpOnly cookies (managed by BFF layer in `shared/bff/`)
- User status workflow: pending → approved/rejected (admin approval required)
