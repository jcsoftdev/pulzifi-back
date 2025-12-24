import { getHttpClient } from '@workspace/shared-http'
import type { Workspace, Organization, User } from './types'
import { MAIN_ROUTES, BOTTOM_ROUTES, type RouteConfig } from './routes'

export interface NavigationData {
  mainRoutes: RouteConfig[]
  bottomRoutes: RouteConfig[]
  workspaces: Workspace[]
  organization: Organization | null
  user: User | null
}

export interface NavigationApiResponse {
  workspaces: Workspace[]
  organization: Organization
  user: User
  customRoutes?: RouteConfig[]
}

/**
 * Navigation Service - Handles fetching navigation data from backend
 * Can be extended to support dynamic routes from backend
 */
export class NavigationService {
  /**
   * Fetch workspaces from backend
   */
  static async fetchWorkspaces(): Promise<Workspace[]> {
    try {
      const http = await getHttpClient()
      const response = await http.get<{ data: Workspace[] }>('/api/workspaces')
      return response.data ?? []
    } catch {
      return []
    }
  }

  /**
   * Fetch current organization
   */
  static async fetchOrganization(): Promise<Organization | null> {
    try {
      const http = await getHttpClient()
      return await http.get<Organization>('/api/organization/current')
    } catch {
      return null
    }
  }

  /**
   * Fetch current user
   */
  static async fetchUser(): Promise<User | null> {
    try {
      const http = await getHttpClient()
      return await http.get<User>('/api/auth/me')
    } catch {
      return null
    }
  }

  /**
   * Get static routes - no async needed, can be called anywhere
   */
  static getRoutes(): { mainRoutes: RouteConfig[]; bottomRoutes: RouteConfig[] } {
    return {
      mainRoutes: MAIN_ROUTES,
      bottomRoutes: BOTTOM_ROUTES,
    }
  }

  /**
   * Get complete navigation data - combines static routes with dynamic data
   */
  static async getNavigationData(): Promise<NavigationData> {
    const [workspaces, organization, user] = await Promise.all([
      this.fetchWorkspaces(),
      this.fetchOrganization(),
      this.fetchUser(),
    ])

    return {
      mainRoutes: MAIN_ROUTES,
      bottomRoutes: BOTTOM_ROUTES,
      workspaces,
      organization,
      user,
    }
  }
}
