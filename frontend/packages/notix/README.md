# @workspace/notix

A headless, trigger-aware toast notification library for React 19+. Zero animation dependencies, pure CSS + Web Animations API, with full Tailwind customizability.

## Features

- **Trigger-Aware Animations**: Morph, fly, or slide animations originating from the trigger element
- **Zero Runtime Dependencies**: Pure CSS custom properties + Web Animations API
- **Fully Headless**: Built-in styled toast or bring your own render function
- **DDD Architecture**: Clean separation of domain, application, and infrastructure layers
- **Tailwind Friendly**: Full CSS customization via className props and CSS variables
- **Spring Physics**: Smooth, natural animations via `linear()` easing curves
- **Auto-Dismiss & Autopilot**: Automatic expansion/collapse with configurable timings
- **Swipe to Dismiss**: Gesture support on mobile
- **Promise API**: Handle async operations with loading → success/error states

## Installation

```bash
bun add @workspace/notix
```

Import styles in your app:

```typescript
import '@workspace/notix/styles.css'
```

## Quick Start

### 1. Add the Toaster

Place the `<Toaster>` at the root of your app (typically in `layout.tsx` or a provider):

```typescript
import { Toaster } from '@workspace/notix'

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <Toaster position="top-right" />
      </body>
    </html>
  )
}
```

### 2. Use the Imperative API

```typescript
import { notix } from '@workspace/notix'

// Basic notifications
notix.success({ title: 'Saved!', description: 'Your changes have been saved.' })
notix.error({ title: 'Error', description: 'Something went wrong.' })
notix.warning({ title: 'Warning', description: 'Please review this.' })
notix.info({ title: 'Info', description: 'Here is some information.' })
notix.loading({ title: 'Loading...' })
notix.action({ title: 'Action', description: 'An action was performed.' })

// Dismiss a specific toast
const id = notix.success({ title: 'Saved' })
notix.dismiss(id)

// Clear all toasts
notix.clear()
notix.clear('top-right') // Clear a specific position
```

### 3. Use the Hook API

```typescript
import { useToast } from '@workspace/notix'

export function MyComponent() {
  const { toasts, show, success, error, dismiss } = useToast()

  return (
    <button onClick={() => success({ title: 'Done!' })}>
      Show Toast
    </button>
  )
}
```

### 4. Trigger-Aware Animations

Animate toast entry/exit from the trigger element's position:

```typescript
import { NotixTrigger, notix } from '@workspace/notix'

export function SaveButton() {
  return (
    <NotixTrigger
      animation="morph"
      toastOptions={{
        title: 'Saved!',
        description: 'Changes saved successfully.',
      }}
    >
      <button>Save</button>
    </NotixTrigger>
  )
}

// Or programmatically with a ref
import { useTriggerRect } from '@workspace/notix'

export function DeleteButton() {
  const { ref, getRect } = useTriggerRect()

  const handleDelete = async () => {
    const triggerRect = getRect()
    await deleteItem()
    notix.success({
      title: 'Deleted',
      animation: 'morph',
      triggerRect,
    })
  }

  return <button ref={ref} onClick={handleDelete}>Delete</button>
}
```

## Animation Modes

Three animation strategies are available:

### `slide` (default)
Slides in from the top with a subtle scale. Does not require a trigger.

```typescript
notix.success({
  title: 'Saved',
  animation: 'slide', // or omit (default)
})
```

### `morph`
Scales and translates from the trigger element's bounding box to the toast's final position. Creates a "morphing" effect.

```typescript
notix.success({
  title: 'Saved',
  animation: 'morph',
  triggerRect: { top: 100, left: 50, width: 80, height: 40, bottom: 140, right: 130 },
})
```

### `fly`
Flies from the trigger element's center to the toast's center with a scale-up effect.

```typescript
notix.success({
  title: 'Saved',
  animation: 'fly',
  triggerRect: { top: 100, left: 50, width: 80, height: 40, bottom: 140, right: 130 },
})
```

## Promise API

Handle async operations with automatic state transitions:

```typescript
import { notix } from '@workspace/notix'

notix.promise(
  fetch('/api/submit').then(r => r.json()),
  {
    loading: { title: 'Submitting...' },
    success: (data) => ({
      title: 'Success!',
      description: `Submitted: ${data.message}`,
    }),
    error: (err) => ({
      title: 'Failed',
      description: err instanceof Error ? err.message : 'Unknown error',
    }),
  }
)
```

## Styling & Customization

### 1. CSS Custom Properties

Override colors and timing globally:

```css
:root {
  --notix-success: #10b981;
  --notix-error: #ef4444;
  --notix-warning: #f59e0b;
  --notix-info: #3b82f6;
  --notix-loading: #6b7280;
  --notix-action: #8b5cf6;

  --notix-duration: 400ms;
  --notix-z-index: 9999;
}
```

### 2. Per-Toast Styling

Override styles for individual toasts via the `styles` object:

```typescript
notix.success({
  title: 'Saved',
  description: 'Your changes have been saved.',
  styles: {
    toast: 'bg-green-50 border border-green-200 rounded-xl',
    title: 'text-green-900 font-semibold',
    description: 'text-green-700',
    badge: 'bg-green-100',
    button: 'bg-green-600 text-white hover:bg-green-700',
  },
})
```

### 3. Tailwind Classes

Pass a `className` to the root toast element:

```typescript
notix.success({
  title: 'Saved',
  className: 'max-w-sm shadow-2xl',
})
```

### 4. Headless Render

Take full control with a custom render function:

```typescript
notix.show({
  render: ({ toast, dismiss, isExpanded, toggle, lifecycle }) => (
    <div className="bg-slate-900 text-white rounded-lg p-4 shadow-xl flex items-center gap-3">
      <span className="flex-1">{toast.title}</span>
      <button
        onClick={dismiss}
        className="text-slate-400 hover:text-slate-200"
      >
        ✕
      </button>
    </div>
  ),
})
```

## Toaster Props

```typescript
<Toaster
  position="top-right"           // 'top-left' | 'top-center' | 'top-right' | 'bottom-left' | 'bottom-center' | 'bottom-right'
  className="custom-class"        // Custom className for viewport
  offset={16}                      // Padding from viewport edge (number or per-side config)
  toastClassName="custom-toast"   // Custom className applied to each toast
  options={{                       // Default options for all toasts
    duration: 6000,
    position: 'top-right',
  }}
/>
```

## Types

```typescript
import type {
  NotixOptions,
  NotixPosition,
  NotixStyles,
  NotixButton,
  NotixPromiseOptions,
  ToastState,
  ToastId,
  ToastData,
  ToastRenderProps,
  ToastLifecycle,
  AnimationMode,
  Duration,
  TriggerRect,
} from '@workspace/notix'
```

## Architecture

Notix is built with Domain-Driven Design principles:

```
Domain Layer
├── entities/          # Toast data models and factories
├── ports/             # IToastStore, IAnimationStrategy interfaces

Application Layer
├── toast-manager.ts   # Use cases (show, success, error, promise, etc.)
└── timer-service.ts   # Auto-dismiss timer management

Infrastructure Layer
├── store/            # ReactiveToastStore (useSyncExternalStore)
├── animation/        # Spring easing, slide, morph, fly strategies
└── styles/           # CSS custom properties and layout

Presentation Layer
├── api.ts            # notix singleton and getGlobalManager()
├── hooks/            # useToast(), useTriggerRect()
└── components/       # Toaster, ToastItem, DefaultToast, NotixTrigger
```

## Accessibility

- Toasts use `aria-live="polite"` on the viewport for screen reader announcements
- Keyboard dismissal supported via Escape key
- Respects `prefers-reduced-motion` for users with motion sensitivity

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+
- Requires Web Animations API and CSS `linear()` easing function support

## Performance

- **Minimal JS**: Animation via Web Animations API (offloaded to browser)
- **No animation libraries**: Zero framer-motion, react-spring, or animation lib overhead
- **Efficient re-renders**: `useSyncExternalStore` batches updates
- **CSS containment**: `contain: layout style` prevents layout recalcs
- **Frozen snapshots**: Store returns frozen arrays for referential equality

## Examples

### With Error Handling

```typescript
const handleSave = async () => {
  try {
    await api.save(data)
    notix.success({ title: 'Saved', description: 'Changes saved successfully.' })
  } catch (err) {
    notix.error({
      title: 'Failed to save',
      description: err instanceof Error ? err.message : 'Unknown error',
    })
  }
}
```

### With Button Action

```typescript
notix.success({
  title: 'Changes saved',
  description: 'View your changes',
  button: {
    title: 'View',
    onClick: () => router.push('/changes'),
  },
})
```

### Dark Mode Support

```typescript
notix.success({
  title: 'Saved',
  styles: {
    toast: 'dark:bg-slate-900 dark:border-slate-700',
    title: 'dark:text-white',
    description: 'dark:text-slate-300',
  },
})
```

## Migration from Sileo

If you're migrating from Sileo, the API is very similar:

```typescript
// Sileo
import { sileo } from 'sileo'
sileo.success({ title: 'Done' })

// Notix
import { notix } from '@workspace/notix'
notix.success({ title: 'Done' })
```

Main differences:
- No `fill` or `roundness` props (use CSS instead)
- No SVG gooey morphing (use morph animation if needed)
- No `<Sileo>` component (use `<Toaster>` instead)

## License

MIT
