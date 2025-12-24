# Features Architecture - DDD + Vertical Slicing

## Structure

Each feature follows this structure:

```
features/
└── [feature-name]/
    ├── domain/
    │   ├── types.ts              # Domain types and interfaces
    │   └── services/             # Domain services (business logic)
    │       └── [feature]-service.ts
    ├── ui/                       # Presentational components
    │   └── [component].tsx
    └── index.tsx                 # Feature entry point
```

## Services Architecture

### Shared Services (`lib/services/`)
**Infrastructure layer** - Services shared across all features:
- `api-client.ts` - Base HTTP client
- `auth.ts` - Authentication & token management
- `storage.ts` - Local storage utilities
- etc.

### Feature Services (`features/[feature]/domain/services/`)
**Domain layer** - Services specific to each feature:
- Contain business logic for that feature
- Use shared services (api-client, auth, etc.)
- Can only be used by their own feature
- Examples:
  - `usage-service.ts` - Usage domain
  - `notification-service.ts` - Notifications domain
  - `workspace-service.ts` - Workspace domain

## Examples

### Using Shared Services
```typescript
// In any feature service
import { WorkspaceApi } from '@workspace/services' 

export class WorkspaceService {
  static async getWorkspaces() {
    return WorkspaceApi.getWorkspaces()
  }
}
```

### Using Feature Services
```typescript
// In layout.tsx (server component)
import { UsageService } from '@/features/usage/domain/services/usage-service'

const checksData = await UsageService.getChecksData()
```

```typescript
// In a page component
import { NotificationService } from '@/features/notifications/domain/services/notification-service'

const notifications = await NotificationService.getNotifications()
```

## Rules

1. **Shared services** can be imported by any feature
2. **Feature services** should only be imported:
   - By their own feature's UI components
   - By layout/pages that need to fetch that feature's data
   - NOT by other features
3. If two features need to share logic, extract it to a shared service
4. Keep domain services focused on business logic, not UI state
