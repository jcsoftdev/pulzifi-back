import { Button } from '@workspace/ui/components/atoms/button'
import Link from 'next/link'
import { NavigationService } from './domain/navigation-service'
import { getBottomRoutes, getMainRoutes } from './domain/routes'
import type { Organization, User, Workspace } from './domain/types'
import { NavigationLink } from './ui/navigation-link'
import { OrganizationSelector } from './ui/organization-selector'
import { ProfileFooter } from './ui/profile-footer'
import { WorkspacesSection } from './ui/workspaces-section'

export interface SidebarFeatureProps {
  organization?: Organization
  user?: User
  workspaces?: Workspace[]
}

/**
 * Server Component - Renders the sidebar structure with routes
 * Fetches all data from backend (organization, user, workspaces)
 * Client components (NavigationLink, WorkspacesSection) handle hydration for interactivity
 */
export async function SidebarFeature({
  organization: providedOrganization,
  user: providedUser,
  workspaces: providedWorkspaces,
}: Readonly<SidebarFeatureProps>) {
  // Fetch all data from backend
  const [organization, user, workspaces] = await Promise.all([
    providedOrganization || NavigationService.fetchOrganization(),
    providedUser || NavigationService.fetchUser(),
    providedWorkspaces || NavigationService.fetchTopWorkspaces(5),
  ])

  const mainRoutes = getMainRoutes()
  const bottomRoutes = getBottomRoutes(user?.role)

  // Split main routes by order - workspaces go after order 1 (Home) and before order 3 (Team)
  const topRoutes = mainRoutes.filter((r) => r.order < 2)
  const afterWorkspacesRoutes = mainRoutes.filter((r) => r.order > 2)

  return (
    <aside className="w-60 h-screen bg-sidebar border-r border-border flex flex-col p-1">
      {/* Logo */}
      <div className="py-2.5 px-3">
        <Button asChild variant="ghost" className="px-1.5 py-1.5 h-auto font-extrabold">
          <Link href="/">
            <span className="text-2xl text-foreground tracking-tight leading-tight">Pulzifi</span>
          </Link>
        </Button>
      </div>

      {/* Organization Selector */}
      <OrganizationSelector organization={organization} />

      {/* Divider */}
      <div className="h-2 border-t border-border mx-3" />

      {/* Navigation */}
      <div className="flex-1 overflow-y-auto py-2 px-2">
        {/* Top Routes (Home) */}
        {topRoutes.map((route) => (
          <NavigationLink key={route.id} route={route} />
        ))}

        {/* Workspaces Section - Client Component for collapse state */}
        <WorkspacesSection workspaces={workspaces || []} />

        {/* Routes after Workspaces (Team) */}
        {afterWorkspacesRoutes.map((route) => (
          <NavigationLink key={route.id} route={route} />
        ))}
      </div>

      {/* Bottom Section */}
      <div className="p-2 space-y-1">
        {/* Bottom Routes (Resources, Settings) */}
        {bottomRoutes.map((route) => (
          <NavigationLink key={route.id} route={route} />
        ))}

        {/* Profile Footer - only if user exists */}
        {user && <ProfileFooter user={user} />}
      </div>
    </aside>
  )
}

export { NavigationService } from './domain/navigation-service'
export {
  BOTTOM_ROUTES,
  getBottomRoutes,
  getMainRoutes,
  MAIN_ROUTES,
  type RouteConfig,
} from './domain/routes'
// Re-export types for convenience
export type { Organization, User, Workspace } from './domain/types'
