import type { LucideIcon } from 'lucide-react'
import { House, Settings, Shapes, ShieldCheck, Users, Workflow } from 'lucide-react'

export type IconName = 'House' | 'Workflow' | 'Users' | 'Shapes' | 'Settings' | 'ShieldCheck'

export interface RouteConfig {
  id: string
  label: string
  href: string
  icon: IconName
  badge?: string | number
  position: 'main' | 'bottom'
  order: number
}

export interface WorkspaceRouteConfig {
  id: string
  baseHref: string
  icon: IconName
  label: string
  position: 'main'
  order: number
  expandable: true
}

// Icon mapping for client components
export const ICON_MAP: Record<IconName, LucideIcon> = {
  House,
  Workflow,
  Users,
  Shapes,
  Settings,
  ShieldCheck,
}

/**
 * Static route definitions - Single source of truth for navigation
 */
export const MAIN_ROUTES: RouteConfig[] = [
  {
    id: 'home',
    label: 'Home',
    href: '/',
    icon: 'House',
    position: 'main',
    order: 1,
  },
  {
    id: 'team',
    label: 'Team',
    href: '/team',
    icon: 'Users',
    position: 'main',
    order: 3,
  },
]

export const WORKSPACES_ROUTE: WorkspaceRouteConfig = {
  id: 'workspaces',
  label: 'Workspaces',
  baseHref: '/workspaces',
  icon: 'Workflow',
  position: 'main',
  order: 2,
  expandable: true,
}

export const BOTTOM_ROUTES: RouteConfig[] = [
  {
    id: 'admin',
    label: 'Admin',
    href: '/admin',
    icon: 'ShieldCheck',
    position: 'bottom',
    order: 0,
  },
  {
    id: 'resources',
    label: 'Resources',
    href: '/resources',
    icon: 'Shapes',
    position: 'bottom',
    order: 2,
  },
  {
    id: 'settings',
    label: 'Settings',
    href: '/settings',
    icon: 'Settings',
    position: 'bottom',
    order: 3,
  },
]

/**
 * Get all routes sorted by order
 */
export function getMainRoutes(): RouteConfig[] {
  return [
    ...MAIN_ROUTES,
  ].sort((a, b) => a.order - b.order)
}

export function getBottomRoutes(userRole?: string): RouteConfig[] {
  return [
    ...BOTTOM_ROUTES,
  ]
    .filter((route) => route.id !== 'admin' || userRole === 'SUPER_ADMIN')
    .sort((a, b) => a.order - b.order)
}

/**
 * Check if a path matches a route
 */
export function isRouteActive(routeHref: string, currentPath: string): boolean {
  if (routeHref === '/') {
    return currentPath === '/'
  }
  return currentPath.startsWith(routeHref)
}

/**
 * Check if a workspace is active
 */
export function isWorkspaceActive(workspaceId: string, currentPath: string): boolean {
  return currentPath.startsWith(`/workspaces/${workspaceId}`)
}
