# Team Module

Organization-level team member management and invitations.

## Domain Entities

- `TeamMember` — org member with denormalized user info (email, name, avatar)

## Use Cases (application/ directories)

- `invite_member` — invite member to organization (also handles resend invite logic)
- `list_members` — list organization members
- `update_member` — update member role
- `remove_member` — remove member from organization

## HTTP Routes (`/team/*`)

- GET `/team/members` — list members
- POST `/team/members` — invite member
- PUT `/team/members/{member_id}` — update member role
- DELETE `/team/members/{member_id}` — remove member
- POST `/team/members/{member_id}/resend-invite` — resend invitation email

## Infrastructure

- PostgreSQL: `organization_members` table (public schema)
- Email: invitation and resend emails via Resend provider
- Invitation status: pending/active

## Notes

- `resend_invite` does not have its own use case directory; resend logic is handled within `invite_member/` or inline in the HTTP module
- Use cases `invite_member`, `remove_member`, and `update_member` each contain their own `errors.go` for use-case-specific error definitions

## Cross-Module Dependencies (violations)

- Imports `modules/auth/infrastructure/middleware` (UserIDKey, UserEmailKey)
- Imports `modules/email/domain/services` (EmailProvider)
- Imports `modules/email/infrastructure/templates` (TeamInvite)

**Recommended:** Define email sending interface in this module's domain. Inject implementation from `cmd/server/modules.go`.
