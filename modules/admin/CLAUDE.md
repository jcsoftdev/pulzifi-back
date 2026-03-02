# Admin Module

## Responsibility

Super-admin user approval/rejection workflow. New user registrations require admin approval before the user can access the platform and have their organization created.

## Entities

- **RegistrationRequest** — ID, UserID, OrganizationName, OrganizationSubdomain, Status (pending/approved/rejected), ReviewedBy, ReviewedAt

## Repository Interfaces

- `RegistrationRequestRepository` — Create, GetByID, GetByUserID, ListPending, UpdateStatus, ExistsPendingBySubdomain

## Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/admin/users/pending` | List pending registration requests |
| PUT | `/admin/users/{id}/approve` | Approve user (creates organization + tenant schema) |
| PUT | `/admin/users/{id}/reject` | Reject user (sends rejection email) |

## Dependencies

- Auth module (user status updates)
- Organization module (creates org on approval)
- Email module (sends approval/rejection emails)

## Constraints

- Requires SUPER_ADMIN permission
- Approval triggers organization creation, which triggers tenant schema creation
- Subdomain uniqueness checked before approval
