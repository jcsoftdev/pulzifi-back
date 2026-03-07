# Workspace Module

Workspace management within organizations.

## Domain Entities

- `Workspace` — workspace with name, type, tags
- `WorkspaceMember` — workspace-level member with role

## Domain Value Objects

- `WorkspaceRole` — immutable workspace role type (`workspace_role.go`, `workspace_role_test.go`)

## Use Cases (application/ directories)

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

- POST `/workspaces` — create workspace
- GET `/workspaces` — list workspaces
- GET `/workspaces/{id}` — get workspace details
- PUT `/workspaces/{id}` — update workspace
- DELETE `/workspaces/{id}` — delete workspace
- GET `/workspaces/{id}/members` — list workspace members
- POST `/workspaces/{id}/members` — add member
- PUT `/workspaces/{id}/members/{member_id}` — update member role
- DELETE `/workspaces/{id}/members/{member_id}` — remove member

## Domain Services

- `WorkspaceAuthorizationService` — workspace-level access control (`workspace_authorization_service.go`, `workspace_authorization_service_test.go`)

## Infrastructure

- PostgreSQL: `workspaces`, `workspace_members` tables (tenant-scoped)
- Authorization middleware: `infrastructure/middleware/workspace_authorization.go`

## Notes

- Most complete hexagonal domain model (entities, errors, repositories, services, value_objects)
- Has the most use cases (9) of any module
- Hierarchy: Organization > Workspace > Pages
- Role-based member permissions per workspace

## Architecture Improvements

### Reference Implementation
This module is the best example of the hexagonal architecture in the codebase. Other modules (especially `usage`, `report`, `auth`) should be refactored to follow this module's patterns:
- Dedicated use case directories with handler/request/response
- Domain value objects with validation
- Domain services for authorization logic
- Infrastructure-level middleware for cross-cutting concerns

### Workspace-Level Quotas
Consider adding workspace-level usage quotas (max pages, max checks per page) in coordination with the `usage` module to enforce plan limits at a more granular level.
