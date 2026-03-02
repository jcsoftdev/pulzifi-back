# Workspace Module

## Responsibility

Workspace CRUD within a tenant, member role-based authorization (Owner/Editor/Viewer), and workspace membership management.

## Entities

- **Workspace** — ID, Name, Type, Tags, CreatedBy, CreatedAt, UpdatedAt, DeletedAt
- **WorkspaceMember** — WorkspaceID, UserID, Role (Owner/Editor/Viewer), InvitedBy, InvitedAt

## Repository Interfaces

- `WorkspaceRepository` — Create, GetByID, List, ListByCreator, Update, Delete, AddMember, GetMember, ListMembers, ListByMember, UpdateMemberRole, RemoveMember

## Routes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/workspaces` | Create workspace |
| GET | `/workspaces` | List workspaces |
| GET | `/workspaces/{id}` | Get workspace |
| PUT | `/workspaces/{id}` | Update workspace |
| DELETE | `/workspaces/{id}` | Delete workspace |
| POST | `/workspaces/{id}/members` | Add member |
| GET | `/workspaces/{id}/members` | List members |
| PUT | `/workspaces/{id}/members/{user_id}` | Update member role |
| DELETE | `/workspaces/{id}/members/{user_id}` | Remove member |

## Dependencies

- Auth middleware (multi-level: global permission → workspace membership → workspace role)

## Constraints

- Tenant-scoped: all workspace data lives in the tenant's PostgreSQL schema
- Creator automatically becomes Owner
- Authorization checks at workspace level via WorkspaceAuthorizationMiddleware
