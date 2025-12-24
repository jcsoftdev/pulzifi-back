# Frontend - Bun + Turbo + Biome Monorepo

Modern monorepo with Bun (package manager), Turbo (task orchestration), and Biome (linting/formatting).

## Quick Start

### Install dependencies

```bash
bun install
```

### Development

```bash
turbo dev
```

### Build

```bash
turbo build
```

### Type Check

```bash
turbo type-check
```

### Linting & Formatting

```bash
bun lint        # Run Biome linter
bun lint:fix    # Fix lint errors
bun format      # Format code with Biome
```

## Structure

```
frontend/
├── apps/
│   └── web/            # Next.js application
├── packages/
│   ├── typescript-config/  # Shared TypeScript configs
│   └── ui/             # Shared UI components (shadcn/ui)
├── biome.json          # Root Biome configuration
└── turbo.json          # Turbo task configuration
```

## Technology Stack

- **Package Manager**: Bun - Fast, modern package manager
- **Task Runner**: Turbo - Optimized monorepo orchestration with caching
- **Build Tool**: Turbopack (Next.js) - Next-gen bundler
- **Framework**: Next.js 16 + React 19
- **Styling**: Tailwind CSS 4
- **Linter/Formatter**: Biome 2.2.0 - Fast alternative to ESLint + Prettier
- **Language**: TypeScript

## How It Works

1. **Bun** manages dependencies and workspaces
2. **Turbo** orchestrates tasks across workspaces (dev, build, type-check)
3. **Biome** handles linting and formatting

## Benefits

✅ **Fast** - Biome is 10x faster than ESLint+Prettier  
✅ **Cached** - Turbo caches builds and outputs  
✅ **Parallel** - Tasks run in parallel automatically  
✅ **Single Config** - One biome.json for all workspaces

## Migration from pnpm

Changed from:
- ❌ pnpm → ✅ Bun
- ❌ ESLint + Prettier → ✅ Biome 2.2.0  
- ✅ Turbo (kept for orchestration)

## Tailwind

Your `tailwind.config.ts` and `globals.css` are already set up to use the components from the `ui` package.

## Using components

To use the components in your app, import them from the `ui` package.

```tsx
import { Button } from "@workspace/ui/components/button"
```
