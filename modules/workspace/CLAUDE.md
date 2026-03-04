# Workspace Module

Workspace management within organizations.

## Domain Entities

- `Workspace` — workspace with name, type, tags
- `WorkspaceMember` — workspace-level member with role

## Use Cases

- `create_workspace` — create workspace in organization
- `list_workspaces` — list workspaces
- `get_workspace` — get workspace details
- `update_workspace` — update workspace
- `delete_workspace` — delete workspace
- `add_workspace_member` — add member to workspace
- `list_workspace_members` — list workspace members
- `remove_workspace_member` — remove member
- `update_member_role` — update member workspace role

## HTTP Routes (`/workspaces/*`, tenant-aware)

- POST `/workspaces`
- GET `/workspaces`
- GET `/workspaces/{id}`
- PUT `/workspaces/{id}`
- DELETE `/workspaces/{id}`
- GET `/workspaces/{id}/members`
- POST `/workspaces/{id}/members`
- PUT `/workspaces/{id}/members/{member_id}`
- DELETE `/workspaces/{id}/members/{member_id}`

## Domain Services

- `WorkspaceAuthorizationService` — workspace-level access control

## Infrastructure

- PostgreSQL: `workspaces`, `workspace_members` tables (tenant-scoped)
- Authorization middleware for workspace-level access

## Notes

- Hierarchy: Organization > Workspace > Pages
- Role-based member permissions per workspace
