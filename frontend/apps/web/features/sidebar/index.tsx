import Link from 'next/link'
import { OrganizationSelector } from './ui/organization-selector'
import { NavigationLink } from './ui/navigation-link'
import { WorkspacesSection } from './ui/workspaces-section'
import { ProfileFooter } from './ui/profile-footer'
import { getMainRoutes, getBottomRoutes } from './domain/routes'
import type { Organization, User, Workspace } from './domain/types'

export interface SidebarFeatureProps {
  organization?: Organization
  user?: User
  workspaces?: Workspace[]
}

/**
 * Server Component - Renders the sidebar structure with routes
 * Client components (NavigationLink, WorkspacesSection) handle hydration for interactivity
 */
export function SidebarFeature({
  organization = {
    id: '1',
    name: 'Dania Morales',
    company: 'Volkswagen INC',
  },
  user = {
    id: '1',
    name: 'Dania Morales',
    role: 'ADMIN',
  },
  workspaces = [
    { id: '1', name: 'Toyota', type: 'Competitor' },
    { id: '2', name: 'Jeep', type: 'Competitor' },
    { id: '3', name: 'Nissan', type: 'Competitor' },
  ],
}: Readonly<SidebarFeatureProps>) {
  const mainRoutes = getMainRoutes()
  const bottomRoutes = getBottomRoutes()

  // Split main routes by order - workspaces go after order 1 (Home) and before order 3 (Team)
  const topRoutes = mainRoutes.filter((r) => r.order < 2)
  const afterWorkspacesRoutes = mainRoutes.filter((r) => r.order > 2)

  return (
    <aside className="w-[229px] h-screen bg-sidebar border-r border-border flex flex-col p-1">
      {/* Logo */}
      <div className="py-2.5 px-3">
        <Link
          href="/"
          className="flex items-center gap-2 px-1.5 py-1.5 hover:bg-muted rounded-lg transition-colors"
        >
          <span className="font-extrabold text-[22.5px] text-foreground tracking-[1%] leading-tight">
            Pulzifi
          </span>
        </Link>
      </div>

      {/* Organization Selector */}
      <OrganizationSelector organization={organization} />

      {/* Divider */}
      <div className="h-[7px] border-t border-border mx-3" />

      {/* Navigation */}
      <div className="flex-1 overflow-y-auto py-2 px-2">
        {/* Top Routes (Home) */}
        {topRoutes.map((route) => (
          <NavigationLink key={route.id} route={route} />
        ))}

        {/* Workspaces Section - Client Component for collapse state */}
        <WorkspacesSection workspaces={workspaces} />

        {/* Routes after Workspaces (Team) */}
        {afterWorkspacesRoutes.map((route) => (
          <NavigationLink key={route.id} route={route} />
        ))}
      </div>

      {/* Bottom Section */}
      <div className="p-2">
        {/* Bottom Routes (Resources, Settings) */}
        {bottomRoutes.map((route) => (
          <NavigationLink key={route.id} route={route} />
        ))}

        {/* Profile Footer */}
        <ProfileFooter
          user={user}
          onLogout={() => console.log('Logout')}
          onSettings={() => console.log('Settings')}
        />
      </div>
    </aside>
  )
}

// Re-export types for convenience
export type { Organization, User, Workspace } from './domain/types'
export { NavigationService } from './domain/navigation-service'
export { getMainRoutes, getBottomRoutes, MAIN_ROUTES, BOTTOM_ROUTES, type RouteConfig } from './domain/routes'
