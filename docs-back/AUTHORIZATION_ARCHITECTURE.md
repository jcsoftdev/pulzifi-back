# Authorization Architecture

## Summary

Pulzifi implements a **two-level authorization system** following Domain-Driven Design (DDD) and Bounded Contexts:

- **Level 1 (Global)**: System permissions — Can the user use this functionality?
- **Level 2 (Domain)**: Workspace roles — What can the user do in THIS specific workspace?

---

## Level 1: Global Authorization (Auth Context)

**Location:** `modules/auth/`
**Schema:** `public` (shared across all tenants)
**Purpose:** Control access to system-level functionalities

### Components

```sql
public.users                  -- System users
public.roles                  -- Global roles (ADMIN, USER, VIEWER)
public.permissions            -- System permissions (workspaces:read, pages:write, etc.)
public.role_permissions       -- Role-permission mappings
public.user_roles             -- Role assignments to users
```

### Predefined Global Roles

| Role | Description | Permissions |
|------|-------------|-------------|
| **ADMIN** | Full administrator | All system permissions |
| **USER** | Standard user | Read + Write on workspaces, pages, alerts |
| **VIEWER** | Read only | Only read permissions |

### Global Permissions

```
workspaces:read       -- Can view workspaces
workspaces:write      -- Can create/edit workspaces
workspaces:delete     -- Can delete workspaces
pages:read            -- Can view pages
pages:write           -- Can create/edit pages
pages:delete          -- Can delete pages
monitoring:read       -- Can view monitoring
monitoring:write      -- Can configure monitoring
alerts:read           -- Can view alerts
alerts:write          -- Can create alerts
reports:read          -- Can view reports
reports:write         -- Can generate reports
users:read            -- Can view users
users:write           -- Can create/edit users
users:delete          -- Can delete users
organizations:read    -- Can view organizations
organizations:write   -- Can manage organizations
```

### Middleware

```go
middleware.AuthMiddleware.Authenticate                        // Validates JWT
middleware.AuthMiddleware.RequirePermission("workspaces", "read")  // Validates global permission
middleware.AuthMiddleware.RequireRole("ADMIN")                // Validates global role
```

---

## Level 2: Domain Authorization (Workspace Context)

**Location:** `modules/workspace/`
**Schema:** `{tenant}` (per-organization)
**Purpose:** Control access to specific workspaces within the domain

### Components

```sql
{tenant}.workspaces           -- Tenant workspaces
{tenant}.workspace_members    -- Members with specific roles
```

### Workspace Roles (Domain Concepts)

| Role | Description | Permissions |
|------|-------------|-------------|
| **owner** | Workspace owner | Create, read, edit, delete, manage members |
| **editor** | Editor | Create, read, edit (NOT delete, NOT manage members) |
| **viewer** | Observer | Read only (NOT modify anything) |

### Value Object

```go
// modules/workspace/domain/value_objects/workspace_role.go
type WorkspaceRole string

const (
    RoleOwner  WorkspaceRole = "owner"
    RoleEditor WorkspaceRole = "editor"
    RoleViewer WorkspaceRole = "viewer"
)

func (r WorkspaceRole) CanRead() bool          { return true }
func (r WorkspaceRole) CanWrite() bool         { return r == RoleOwner || r == RoleEditor }
func (r WorkspaceRole) CanDelete() bool        { return r == RoleOwner }
func (r WorkspaceRole) CanManageMembers() bool { return r == RoleOwner }
```

### Domain Service

```go
// modules/workspace/domain/services/workspace_authorization_service.go
type WorkspaceAuthorizationService struct {
    memberRepo repositories.WorkspaceMemberRepository
}

func (s *WorkspaceAuthorizationService) CanReadWorkspace(ctx, tenant, workspaceID, userID)
func (s *WorkspaceAuthorizationService) CanWriteWorkspace(ctx, tenant, workspaceID, userID)
func (s *WorkspaceAuthorizationService) CanDeleteWorkspace(ctx, tenant, workspaceID, userID)
func (s *WorkspaceAuthorizationService) CanManageMembers(ctx, tenant, workspaceID, userID)
```

---

## Full Authorization Flow

### Example: User attempts to delete a workspace

```
1. Request: DELETE /api/v1/workspaces/{id}
   |
2. [Auth Middleware] Validate JWT
   OK: Token valid -> userID = "123", roles = ["USER"], permissions = ["workspaces:delete"]
   |
3. [Organization Middleware] Validate org membership (tenant)
   OK: User is a member of the organization
   |
4. [LEVEL 1] AuthMiddleware.RequirePermission("workspaces", "delete")
   OK: User has global permission "workspaces:delete"
   |
5. [LEVEL 2] WorkspaceAuthMiddleware.RequireWorkspaceMembership
   OK: User is a member of the workspace
   |
6. [LEVEL 2] WorkspaceAuthMiddleware.RequireWorkspaceRole(RoleOwner)
   OK: User has role "owner" in the workspace
   |
7. [Handler] Execute business logic
   OK: Workspace deleted
```

### Rejection: User without global permission

```
1. Request: DELETE /api/v1/workspaces/{id}
   |
2. [Auth Middleware] Validate JWT
   OK: Token valid -> userID = "456", roles = ["VIEWER"], permissions = ["workspaces:read"]
   |
3. [Organization Middleware] Validate organization
   OK: User is a member
   |
4. [LEVEL 1] AuthMiddleware.RequirePermission("workspaces", "delete")
   FAIL: User does NOT have "workspaces:delete"
   |
   Response: 403 Forbidden - "insufficient permissions"

   (Never reaches Level 2)
```

### Rejection: User without adequate workspace role

```
1. Request: DELETE /api/v1/workspaces/{id}
   |
2-4. [Passes Level 1] User has global permission
   |
5. [LEVEL 2] WorkspaceAuthMiddleware.RequireWorkspaceMembership
   OK: User is a member of the workspace
   |
6. [LEVEL 2] WorkspaceAuthMiddleware.RequireWorkspaceRole(RoleOwner)
   FAIL: User has role "editor" (NOT "owner")
   |
   Response: 403 Forbidden - "insufficient permissions in this workspace"
```

---

## Permission Matrix by Endpoint

### Workspace Endpoints

| Endpoint | HTTP | Level 1 (Global) | Level 2 (Role) |
|----------|------|-----------------|----------------|
| Create workspace | `POST /workspaces` | `workspaces:write` | N/A (auto-owner) |
| List workspaces | `GET /workspaces` | `workspaces:read` | N/A (auto-filtered) |
| View workspace | `GET /workspaces/{id}` | `workspaces:read` | `viewer` (minimum) |
| Edit workspace | `PUT /workspaces/{id}` | `workspaces:write` | `editor` (minimum) |
| Delete workspace | `DELETE /workspaces/{id}` | `workspaces:delete` | `owner` (only) |
| Add member | `POST /workspaces/{id}/members` | `workspaces:write` | `owner` (only) |
| List members | `GET /workspaces/{id}/members` | `workspaces:read` | `viewer` (minimum) |
| Remove member | `DELETE /workspaces/{id}/members/{user_id}` | `workspaces:write` | `owner` (only) |

---

## Router Implementation

```go
// modules/workspace/infrastructure/http/module.go
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
    workspaceAuth := workspacemw.NewWorkspaceAuthorizationMiddleware(m.db)

    router.Route("/workspaces", func(r chi.Router) {
        // ========================================
        // LEVEL 1: Authentication and Global Permissions
        // ========================================
        r.Use(middleware.AuthMiddleware.Authenticate)
        r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)

        // Public endpoints (Level 1 only)
        r.Group(func(r chi.Router) {
            r.Use(middleware.AuthMiddleware.RequirePermission("workspaces", "write"))
            r.Post("/", m.handleCreateWorkspace)
        })

        // ========================================
        // LEVEL 2: Workspace Authorization
        // ========================================
        // View workspace (viewer+)
        r.Group(func(r chi.Router) {
            r.Use(middleware.AuthMiddleware.RequirePermission("workspaces", "read"))
            r.Use(workspaceAuth.RequireWorkspaceMembership)
            r.Get("/{id}", m.handleGetWorkspace)
        })

        // Edit workspace (editor+)
        r.Group(func(r chi.Router) {
            r.Use(middleware.AuthMiddleware.RequirePermission("workspaces", "write"))
            r.Use(workspaceAuth.RequireWorkspaceMembership)
            r.Use(workspaceAuth.RequireWorkspaceRole(value_objects.RoleEditor))
            r.Put("/{id}", m.handleUpdateWorkspace)
        })

        // Delete workspace (owner only)
        r.Group(func(r chi.Router) {
            r.Use(middleware.AuthMiddleware.RequirePermission("workspaces", "delete"))
            r.Use(workspaceAuth.RequireWorkspaceMembership)
            r.Use(workspaceAuth.RequireWorkspaceRole(value_objects.RoleOwner))
            r.Delete("/{id}", m.handleDeleteWorkspace)
        })
    })
}
```

---

## DDD Principles Applied

### Bounded Contexts

```
┌─────────────────────────┐     ┌──────────────────────────┐
│   Auth Context          │     │  Workspace Context       │
│   (Global System)       │     │  (Domain-Specific)       │
├─────────────────────────┤     ├──────────────────────────┤
│ - Roles (ADMIN, USER)   │     │ - Roles (owner, editor)  │
│ - Permissions (read,..) │     │ - Business logic         │
│ - Authentication        │     │ - Workspace rules        │
└─────────────────────────┘     └──────────────────────────┘
         |                                  |
    System-Wide                      Domain-Specific
```

### Separation of Concerns

| Responsibility | Auth Module | Workspace Module |
|----------------|-------------|------------------|
| Authentication (JWT) | Yes | - |
| Global permissions | Yes | - |
| System roles | Yes | - |
| Workspace rules | - | Yes |
| Membership | - | Yes |
| Business logic | - | Yes |

### No Coupling Between Modules

```go
// BAD (coupling)
// Workspace module imports Auth module to validate permissions
import "modules/auth/domain/services"
authService.HasPermission(...)

// GOOD (independence)
// Workspace only validates its own domain rules
member.Role.CanDelete()  // Internal domain logic
```

---

## Adding New Endpoints

1. Define global permission in `public.permissions` migration
2. Assign permission to roles in `public.role_permissions`
3. Apply Level 1 middleware in router
4. Apply Level 2 middleware if resource-specific
5. Document in the permission matrix above

### Adding New Modules

Each module can have its own domain roles (like workspace has owner/editor/viewer), but must always:
1. Validate global permissions first (Level 1)
2. Validate domain rules after (Level 2)
3. Maintain bounded context separation
