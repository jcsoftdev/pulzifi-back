# Team Module

## Responsibility

Organization-level team member management, invitation workflow, role assignment, and invitation resending.

## Entities

- **TeamMember** — ID, OrganizationID, UserID, Role, InvitedBy, JoinedAt, InvitationStatus, FirstName, LastName, Email, AvatarURL

## Repository Interfaces

- `TeamMemberRepository` — ListByOrganization, GetByID, GetByUserAndOrg, FindUserByEmail, CreateUser, AddMember, UpdateRole, UpdateInvitationStatus, Remove, GetOrganizationIDBySubdomain

## Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/team/members` | List team members |
| POST | `/team/members` | Invite member |
| PUT | `/team/members/{member_id}` | Update member role |
| DELETE | `/team/members/{member_id}` | Remove member |
| POST | `/team/members/{member_id}/resend-invite` | Resend invitation email |

## Dependencies

- Auth module (creates user account for invited member)
- Email module (sends invitation emails with password reset token)

## Constraints

- Tenant-scoped
- Invitation creates a user with pending status and a password reset token
- Accepting invitation activates the membership via password reset flow
