# Frontend Development Guidelines - Pulzifi

## Architecture Overview

This monorepo uses:
- **Bun** - Package manager + workspaces
- **Turbo** - Task orchestration with caching
- **Biome** - Linting and formatting
- **Next.js 16** - React framework
- **TypeScript** - Type safety

---

## Workspace Structure

```
frontend/
├── apps/
│   └── web/              # Next.js application (DDD + Vertical Slicing)
├── packages/
│   ├── ui/              # Atomic Design System
│   └── typescript-config/
```

---

## Package: UI (Atomic Design)

**Location:** `packages/ui/src/components/`

### Structure

```
components/
├── atoms/               # Basic building blocks
│   ├── button.tsx
│   ├── badge.tsx
│   ├── icon-button.tsx
│   └── avatar.tsx
├── molecules/           # Simple component combinations
│   ├── notification-button.tsx
│   ├── checks-tag.tsx
│   └── user-profile.tsx
└── organisms/           # Complex UI sections
    └── header.tsx
```

### Atomic Design Rules

**Atoms:**
- Single responsibility
- No business logic
- Highly reusable
- Props-driven styling
- No external dependencies (except lucide-react for icons)

**Molecules:**
- Combine 2-3 atoms
- Simple interaction logic
- Reusable across features
- Can use hooks for UI state only

**Organisms:**
- Complex UI sections
- Combine molecules and atoms
- Can have internal state
- Feature-agnostic

### Example: Creating a New Atom

```tsx
// packages/ui/src/components/atoms/input.tsx
import * as React from "react"
import { cn } from "../../lib/utils"

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  error?: boolean
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, error, ...props }, ref) => {
    return (
      <input
        className={cn(
          "px-3 py-2 border rounded-md",
          error && "border-red-500",
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)
Input.displayName = "Input"

export { Input }
```

### Export Pattern

Always export from `packages/ui/src/components/index.ts`:

```tsx
// Atoms
export { Button } from "./atoms/button"
export { Input } from "./atoms/input"

// Molecules
export { NotificationButton } from "./molecules/notification-button"

// Organisms
export { Header } from "./organisms/header"
```

---

## App: Web (DDD + Vertical Slicing + Feature-Based)

**Location:** `apps/web/src/features/`

### Structure

```
features/
├── dashboard/
│   ├── domain/          # Business logic & types
│   │   └── types.ts
│   ├── ui/             # Presentational components
│   │   ├── dashboard-header.tsx
│   │   ├── stat-card.tsx
│   │   └── empty-state-card.tsx
│   └── index.tsx       # Feature entry point
├── sidebar/
│   ├── domain/
│   │   └── types.ts
│   ├── ui/
│   │   ├── navigation-item.tsx
│   │   ├── organization-selector.tsx
│   │   └── profile-footer.tsx
│   └── index.tsx
└── [feature-name]/
```

### Feature Architecture

#### 1. Domain Layer (`domain/`)
- **Purpose:** Business logic, types, interfaces
- **Contains:** TypeScript types, interfaces, validators
- **Rules:**
  - No UI components
  - No framework-specific code
  - Pure TypeScript/JavaScript
  - Can import from other domain layers

```tsx
// features/workspace/domain/types.ts
export interface Workspace {
  id: string
  name: string
  type: 'Personal' | 'Team' | 'Competitor'
  pages: Page[]
}

export interface WorkspaceRepository {
  findAll(): Promise<Workspace[]>
  findById(id: string): Promise<Workspace>
  create(data: CreateWorkspaceDto): Promise<Workspace>
}
```

#### 2. UI Layer (`ui/`)
- **Purpose:** Presentational components specific to this feature
- **Contains:** React components that are NOT atomic (feature-specific)
- **Rules:**
  - Can import from `@workspace/ui` (atomic components)
  - Can import from own `domain/`
  - Props-driven
  - Minimal logic (presentation only)

```tsx
// features/workspace/ui/workspace-card.tsx
import { Card } from "@workspace/ui/components/atoms/card"
import { Badge } from "@workspace/ui/components/atoms/badge"
import type { Workspace } from "../domain/types"

export interface WorkspaceCardProps {
  workspace: Workspace
  onSelect: (id: string) => void
}

export function WorkspaceCard({ workspace, onSelect }: WorkspaceCardProps) {
  return (
    <Card onClick={() => onSelect(workspace.id)}>
      <h3>{workspace.name}</h3>
      <Badge>{workspace.type}</Badge>
      <p>{workspace.pages.length} pages</p>
    </Card>
  )
}
```

#### 3. Feature Entry (`index.tsx`)
- **Purpose:** Orchestrates the feature (like a controller)
- **Contains:** Main feature component with hooks, API calls, state
- **Rules:**
  - Can use React hooks
  - Handles side effects (API calls)
  - Manages feature state
  - Composes UI components

```tsx
// features/workspace/index.tsx
'use client'

import { useState, useEffect } from 'react'
import { WorkspaceCard } from './ui/workspace-card'
import type { Workspace } from './domain/types'

export function WorkspaceFeature() {
  const [workspaces, setWorkspaces] = useState<Workspace[]>([])
  
  useEffect(() => {
    // Fetch workspaces from API
  }, [])

  return (
    <div>
      {workspaces.map(workspace => (
        <WorkspaceCard 
          key={workspace.id} 
          workspace={workspace}
          onSelect={(id) => console.log(id)}
        />
      ))}
    </div>
  )
}
```

### When to Create a New Feature

Create a new feature when:
- ✅ It represents a distinct business capability
- ✅ It has its own domain models
- ✅ It can be developed/tested independently
- ✅ Multiple pages might use it

Examples:
- `dashboard/` - Home dashboard
- `workspace/` - Workspace management
- `monitoring/` - Page monitoring
- `auth/` - Authentication

### Component Decision Tree

```
Is it a basic HTML element styled?
├─ YES → Atom in packages/ui
└─ NO
   Is it reusable across multiple features?
   ├─ YES → Molecule/Organism in packages/ui
   └─ NO → UI component in feature/ui/
```

---

## Import Rules

### ✅ Allowed Imports

**In packages/ui:**
```tsx
import { cn } from "../../lib/utils"
import { Button } from "../atoms/button"
import { Badge } from "../atoms/badge"
```

**In apps/web features:**
```tsx
import { Button } from "@workspace/ui/components/button"
import { Header } from "@workspace/ui/components/organisms/header"
import type { Workspace } from "./domain/types"
import { WorkspaceCard } from "./ui/workspace-card"
```

### ❌ Prohibited Imports

- ❌ Never import features into `packages/ui`
- ❌ Never import UI components directly across features
- ❌ Never import from `node_modules` in domain layer
- ❌ Never import entire `@workspace/ui` (use specific exports)

---

## Naming Conventions

### Files
- Components: `PascalCase.tsx` (e.g., `WorkspaceCard.tsx`)
- Types: `kebab-case.ts` (e.g., `workspace-types.ts`)
- Utils: `kebab-case.ts` (e.g., `format-date.ts`)

### Components
- Atoms: Simple noun (e.g., `Button`, `Badge`, `Input`)
- Molecules: Descriptive noun (e.g., `NotificationButton`, `UserProfile`)
- Organisms: Section name (e.g., `Header`, `Sidebar`, `Footer`)
- Features: `[Feature]Feature` (e.g., `DashboardFeature`, `WorkspaceFeature`)

### Types
```tsx
// Props
export interface ComponentNameProps { }

// Domain
export interface EntityName { }
export type EntityStatus = 'active' | 'inactive'
```

---

## Styling Guidelines

### Use Tailwind CSS
- Always use Tailwind utility classes
- Use `cn()` helper for conditional classes
- Create variants with `class-variance-authority`

```tsx
import { cva } from "class-variance-authority"
import { cn } from "@workspace/ui" 

const buttonVariants = cva(
  "px-4 py-2 rounded-md font-medium transition-colors",
  {
    variants: {
      variant: {
        primary: "bg-purple-600 text-white hover:bg-purple-700",
        secondary: "bg-gray-200 text-gray-900 hover:bg-gray-300",
      },
      size: {
        sm: "text-sm px-3 py-1.5",
        md: "text-base px-4 py-2",
        lg: "text-lg px-6 py-3",
      }
    },
    defaultVariants: {
      variant: "primary",
      size: "md"
    }
  }
)
```

---

## Commands

```bash
# Development
turbo dev                 # Run all apps in dev mode
cd apps/web && bun dev   # Run specific app

# Build
turbo build              # Build all workspaces

# Linting & Formatting
bun lint                 # Lint with Biome
bun lint:fix             # Fix lint errors
bun format               # Format code

# Type Check
turbo type-check         # Check TypeScript in all workspaces
```

---

## Adding New Dependencies

### To UI package:
```bash
cd packages/ui
bun add [package-name]
```

### To Web app:
```bash
cd apps/web
bun add [package-name]
```

### To Root (shared):
```bash
bun add -D [package-name]
```

---

## Testing Strategy

### Unit Tests
- Test domain logic in `domain/`
- Test pure functions and utilities
- Use Vitest or Jest

### Component Tests
- Test UI components in isolation
- Use React Testing Library
- Test user interactions

### Integration Tests
- Test feature flows
- Mock API calls
- Test state management

---

## Best Practices

### 1. Single Responsibility
Each component should do ONE thing well.

### 2. Composition over Inheritance
Build complex UIs by composing simple components.

### 3. Props Over State
Prefer controlled components with props.

### 4. Type Everything
Use TypeScript for all props, state, and domain models.

### 5. Accessibility
- Use semantic HTML
- Add ARIA labels
- Support keyboard navigation

### 6. Performance
- Use React.memo for expensive renders
- Lazy load features with dynamic imports
- Optimize images with Next.js Image

### 7. Error Boundaries
Wrap features in error boundaries.

```tsx
// apps/web/src/components/error-boundary.tsx
'use client'

import { Component, type ReactNode } from 'react'

export class ErrorBoundary extends Component<
  { children: ReactNode },
  { hasError: boolean }
> {
  state = { hasError: false }

  static getDerivedStateFromError() {
    return { hasError: true }
  }

  render() {
    if (this.state.hasError) {
      return <div>Something went wrong</div>
    }
    return this.props.children
  }
}
```

---

## Quick Reference

| Layer | Location | Purpose | Can Import |
|-------|----------|---------|------------|
| **Atoms** | `packages/ui/atoms/` | Basic elements | Other atoms, utils |
| **Molecules** | `packages/ui/molecules/` | Simple combos | Atoms, utils |
| **Organisms** | `packages/ui/organisms/` | Complex sections | Atoms, molecules, utils |
| **Domain** | `features/*/domain/` | Business logic | Other domain layers |
| **UI** | `features/*/ui/` | Feature UI | Domain, @workspace/ui |
| **Feature** | `features/*/index.tsx` | Orchestration | Domain, UI, hooks, API |

---

## Example: Creating a New Feature

1. **Create structure:**
```bash
mkdir -p apps/web/src/features/workspace/{domain,ui}
touch apps/web/src/features/workspace/{domain/types.ts,ui/.gitkeep,index.tsx}
```

2. **Define domain types:**
```tsx
// domain/types.ts
export interface Workspace {
  id: string
  name: string
}
```

3. **Create UI components:**
```tsx
// ui/workspace-card.tsx
import type { Workspace } from '../domain/types'

export function WorkspaceCard({ workspace }: { workspace: Workspace }) {
  return <div>{workspace.name}</div>
}
```

4. **Compose feature:**
```tsx
// index.tsx
'use client'

import { WorkspaceCard } from './ui/workspace-card'

export function WorkspaceFeature() {
  return <WorkspaceCard workspace={{ id: '1', name: 'Test' }} />
}
```

---

## Resources

- [Atomic Design](https://bradfrost.com/blog/post/atomic-web-design/)
- [Vertical Slice Architecture](https://www.jimmybogard.com/vertical-slice-architecture/)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Bun Docs](https://bun.sh/docs)
- [Turbo Docs](https://turbo.build/repo/docs)
- [Next.js Docs](https://nextjs.org/docs)
