# Frontend

Bun workspace monorepo with Turborepo orchestration for the Pulzifi web application.

## Commands

```bash
bun dev            # Start Next.js dev server on :3001 (via Turborepo)
bun run build      # Production build (via Turborepo)
bun run lint:fix   # Format and lint with Biome
bun run format     # Format with Biome
bun run type-check # TypeScript type checking (via Turborepo)
```

## Technology Stack

- **Runtime:** Bun 1.1.43
- **Framework:** Next.js 16.1.6 (App Router, Turbopack)
- **UI:** React 19.2.4, Tailwind CSS v4, Radix UI primitives
- **Tooling:** Turborepo, Biome (linter + formatter), TypeScript 5.9
- **Fonts:** Geist, Geist Mono, DM Serif Display, Outfit (Google Fonts)

## Monorepo Structure

### Apps
- `apps/web/` — Next.js App Router application (port :3001)

### Packages
| Package | Name | Description |
|---------|------|-------------|
| `packages/ui/` | `@workspace/ui` | Atomic Design component library (atoms/molecules/organisms) using Radix UI + Tailwind |
| `packages/services/` | `@workspace/services` | API client layer — 12 service files (auth, workspace, page, dashboard, notification, organization, team, usage, super-admin, integration, report, monitoring) |
| `packages/shared-http/` | `@workspace/shared-http` | Tenant-aware HTTP factory (Axios for browser, Fetch for SSR), `IHttpClient` interface, subdomain extraction |
| `packages/notix/` | `@workspace/notix` | Toast notification library (hexagonal architecture, motion animations) |
| `packages/typescript-config/` | `@workspace/typescript-config` | Shared TSConfig presets (base, nextjs, react-library) |

## Route Groups (16 pages, 4 layouts)

- `(auth)/` — Login, invite acceptance (redirects if already authenticated)
- `(main)/` — Authenticated app wrapped in `AuthGuard` + `AppShell` (sidebar + header)
  - `/dashboard` — Organization overview statistics
  - `/workspaces` — Workspace listing
  - `/workspaces/[id]` — Workspace detail with nested pages and reports
  - `/workspaces/[id]/pages/[pageId]` — Page detail with checks history
  - `/workspaces/[id]/pages/[pageId]/changes` — Visual change comparison
  - `/workspaces/[id]/reports` — Workspace reports
  - `/settings` — Organization settings and integrations
  - `/team` — Team member management
  - `/admin` — Super admin panel
- `(public)/` — Registration (no auth required)
- `(demo)/` — Experimental pages (lecture-ai)

## Features (17 vertical slices)

Each feature follows `features/{name}/` with UI, application, domain layers:
account-settings, auth, changes-view, dashboard, landing, navigation, notifications, page, page-detail, reports, settings, sidebar, super-admin, team, usage, workspace, workspace-detail

## Auth Protection

- No Next.js middleware file — auth is handled by `AuthGuard` (async React Server Component)
- `AuthGuard` calls `AuthApi.getCurrentUser()` server-side
- On `UnauthorizedError`, renders `SessionRefresher` client component to attempt client-side token refresh
- Users with non-approved status are redirected to `/login?error=PendingApproval`
- `(auth)` layout checks if user is already authenticated and redirects to tenant subdomain

## Multi-Tenant Support

- Subdomains used for tenant routing (e.g., `tenant1.app.local:3001`)
- `@workspace/shared-http` extracts tenant from subdomain for API requests
- `next.config.mjs` allows dev origins: `*.localhost`, `*.app.local`, `*.pulzifi.local`, `*.pulzifi.com`
- `setup-local-domains.sh` configures local domain names for development

## Coding Conventions

- Biome for formatting and linting (spaces, indent width 2, line width 100)
- Atomic Design for `@workspace/ui`: atoms (primitives), molecules (composed), organisms (complex)
- Feature slices follow vertical architecture: UI -> application -> domain
- API clients in `@workspace/services` use `IHttpClient` interface from `@workspace/shared-http`
- Server Components by default; Client Components only when interactivity is needed

## Architecture Improvements

### Dependency Version Mismatch
`lucide-react` has different versions across packages:
- `packages/ui/`: `^0.562.0`
- `apps/web/`: `^0.574.0`

This can cause icon rendering inconsistencies or duplicate bundles. Align to a single version, preferably in the root `package.json` or via Bun workspace resolution.

### API Client Error Handling
The service files in `@workspace/services` should have consistent error handling patterns:
- Standardize error response parsing across all 12 service files
- Consider a shared error interceptor in `@workspace/shared-http` for common error codes (401, 403, 429, 500)

### Type Safety
- Consider generating TypeScript types from the Go API's Swagger/OpenAPI spec to keep frontend and backend types in sync
- This eliminates manual type duplication and drift between frontend DTOs and backend response types
