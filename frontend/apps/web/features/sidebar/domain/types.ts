import type { Organization as OrganizationService, User as UserService } from '@workspace/services'

// Re-export from services (already in camelCase)
export type Organization = OrganizationService
export type User = UserService

export interface Workspace {
  id: string
  name: string
  type: 'Personal' | 'Team' | 'Competitor'
  pageCount?: number
}

export interface NavigationItem {
  id: string
  label: string
  href: string
  icon: React.ReactNode
  active: boolean
  badge?: string | number
}
