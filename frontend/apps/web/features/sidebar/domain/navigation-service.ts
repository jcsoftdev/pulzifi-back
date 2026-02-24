import { AuthApi, OrganizationApi, WorkspaceApi } from '@workspace/services'
import { BOTTOM_ROUTES, MAIN_ROUTES, type RouteConfig } from './routes'
import type { Organization, User, Workspace } from './types'

export interface NavigationData {
  mainRoutes: RouteConfig[]
  bottomRoutes: RouteConfig[]
  workspaces: Workspace[]
  organization: Organization
  user: User
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
export const NavigationService = {
  /**
   * Fetch top N workspaces from backend (uses workspace-api from packages/services)
   */
  async fetchTopWorkspaces(limit: number = 5): Promise<Workspace[]> {
    const response = await WorkspaceApi.listWorkspaces({
      limit,
    })

    // Transform backend DTO to frontend domain type
    return response.workspaces.map((dto) => ({
      id: dto.id,
      name: dto.name,
      type: dto.type as Workspace['type'],
    }))
  },

  /**
   * Fetch all workspaces from backend
   */
  async fetchWorkspaces(): Promise<Workspace[]> {
    const response = await WorkspaceApi.listWorkspaces()

    // Transform backend DTO to frontend domain type
    return response.workspaces.map((dto) => ({
      id: dto.id,
      name: dto.name,
      type: dto.type as Workspace['type'],
    }))
  },

  /**
   * Fetch current organization
   */
  async fetchOrganization(): Promise<Organization> {
    return await OrganizationApi.getCurrentOrganization()
  },

  /**
   * Fetch current user
   */
  async fetchUser(): Promise<User> {
    return await AuthApi.getCurrentUser()
  },

  /**
   * Get static routes - no async needed, can be called anywhere
   */
  getRoutes(): {
    mainRoutes: RouteConfig[]
    bottomRoutes: RouteConfig[]
  } {
    return {
      mainRoutes: MAIN_ROUTES,
      bottomRoutes: BOTTOM_ROUTES,
    }
  },

  /**
   * Get complete navigation data - combines static routes with dynamic data
   */
  async getNavigationData(): Promise<NavigationData> {
    const [workspaces, organization, user] = await Promise.all([
      NavigationService.fetchTopWorkspaces(5),
      NavigationService.fetchOrganization(),
      NavigationService.fetchUser(),
    ])

    return {
      mainRoutes: MAIN_ROUTES,
      bottomRoutes: BOTTOM_ROUTES,
      workspaces,
      organization,
      user,
    }
  },
}
