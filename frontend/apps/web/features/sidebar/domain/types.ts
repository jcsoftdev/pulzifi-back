export interface Organization {
  id: string
  name: string
  company?: string
  avatar?: string
}

export interface User {
  id: string
  name: string
  role: string
  email?: string
  avatar?: string
}

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

