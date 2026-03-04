# Team Module

Organization-level team member management and invitations.

## Domain Entities

- `TeamMember` — org member with denormalized user info (email, name, avatar)

## Use Cases

- `invite_member` — invite member to organization
- `list_members` — list organization members
- `update_member` — update member role
- `remove_member` — remove member from organization
- `resend_invite` — resend invitation email

## HTTP Routes (`/team/*`)

- GET `/team/members`
- POST `/team/members`
- PUT `/team/members/{member_id}`
- DELETE `/team/members/{member_id}`
- POST `/team/members/{member_id}/resend-invite`

## Infrastructure

- PostgreSQL: `organization_members` table (public schema)
- Email: invitation and resend emails
- Invitation status: pending/active
