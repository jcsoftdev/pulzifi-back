# Notification Abstraction Layer

A DDD-based abstraction over the toast notification library, enabling seamless library swaps without touching feature code.

## Architecture

```
Feature Code (no library knowledge)
        ↓
@/lib/notification/index.ts (notification singleton)
        ↓
INotificationPort (interface)
        ↓
NotixNotificationAdapter (implementation)
        ↓
@workspace/notix (library)
```

## Quick Start

### 1. Import the Singleton

```typescript
import { notification } from '@/lib/notification'

// Show a toast
notification.success({
  title: 'Saved',
  description: 'Your changes have been saved.',
})
```

### 2. Basic API

```typescript
// Success notification
notification.success({ title: 'Done!', description: 'Operation completed.' })

// Error notification
notification.error({ title: 'Error', description: 'Something went wrong.' })

// Warning
notification.warning({ title: 'Warning', description: 'Please review.' })

// Info
notification.info({ title: 'Info', description: 'Here is information.' })

// Loading (no auto-dismiss)
notification.loading({ title: 'Processing...' })

// Action
notification.action({ title: 'Action', description: 'Action performed.' })

// Dismiss
const id = notification.success({ title: 'Saved' })
notification.dismiss(id)

// Clear all
notification.clear()
```

## Real-World Examples

### API Call with Error Handling

```typescript
import { notification } from '@/lib/notification'

const handleSave = async () => {
  try {
    await api.updatePage(page.id, data)
    notification.success({
      title: 'Page updated',
      description: `"${page.name}" has been updated.`,
    })
  } catch (err) {
    notification.error({
      title: 'Failed to update page',
      description: err instanceof Error ? err.message : 'Please try again.',
    })
  }
}
```

### Promise-Based Operations

```typescript
notification.promise(
  deleteWorkspace(workspace.id),
  {
    loading: { title: 'Deleting workspace...' },
    success: { title: 'Workspace deleted' },
    error: (err) => ({
      title: 'Failed to delete',
      description: err instanceof Error ? err.message : 'Please try again.',
    }),
  }
)
```

### With Button Action

```typescript
notification.success({
  title: 'Changes saved',
  description: 'View your recent changes',
  level: 'success',
  button: {
    title: 'View Changes',
    onClick: () => router.push(`/workspaces/${workspace.id}/pages/${page.id}/changes`),
  },
})
```

### Custom Styling

```typescript
notification.success({
  title: 'Page added',
  description: '"Home" has been added.',
  styles: {
    toast: 'bg-green-900 border border-green-700',
    title: 'text-green-100 font-semibold',
    description: 'text-green-200',
    badge: 'bg-green-500',
  },
})
```

## Types

All notification types are framework-agnostic and located in the abstraction:

```typescript
import type {
  INotificationPort,      // Main interface
  NotificationOptions,    // Options for show/success/error/etc
  NotificationPromiseOptions,  // Options for promise()
  NotificationLevel,      // 'success' | 'error' | 'warning' | 'info' | 'loading' | 'action'
  NotificationId,         // Return type of show/success/error/etc
} from '@/lib/notification'
```

## Swapping Libraries

To swap from notix to another toast library:

1. Create a new adapter class implementing `INotificationPort`:

```typescript
// lib/notification/my-library-adapter.ts
import type { INotificationPort, NotificationOptions } from './notification-port'

export class MyLibraryAdapter implements INotificationPort {
  show(options: NotificationOptions) {
    // Use your library here
    return myLibrary.show({
      title: options.title,
      description: options.description,
    })
  }

  success(options: NotificationOptions) {
    return myLibrary.success(options)
  }

  // ... implement other methods
}
```

2. Update the singleton in `index.ts`:

```typescript
// Change this line:
export const notification: INotificationPort = new NotixNotificationAdapter()

// To this:
export const notification: INotificationPort = new MyLibraryAdapter()
```

3. Update the provider component:

```typescript
// lib/notification/notification-provider.tsx
export function NotificationProvider() {
  return <MyLibraryToaster position="top-right" />
}
```

**That's it.** No feature code changes needed. All 13 feature files continue working unchanged.

## Design Principles

### 1. **Port & Adapter Pattern**
- `INotificationPort`: Framework-agnostic interface
- `NotixNotificationAdapter`: Implementation-specific adapter
- Feature code depends only on the port, not the implementation

### 2. **Dependency Inversion**
- Feature code imports from `@/lib/notification`
- The notification abstraction imports the library
- Feature code **never** imports from `@workspace/notix` directly

### 3. **Single Responsibility**
- Each file has one reason to change:
  - `notification-port.ts` — Only if the API contract changes (rare)
  - `notix-adapter.ts` — Only if notix API changes
  - `notification-provider.tsx` — Only if the Toaster component changes
  - Feature code — Only if business logic changes

## Notes

- The singleton is created once at module load time
- All toasts share the same global store (no per-component state)
- The `<NotificationProvider />` must be rendered in your root layout
- The library-specific imports are **isolated** to the abstraction layer
