export interface Organization {
  id: string
  name: string
  company: string
}

export interface User {
  id: string
  name: string
  role: 'ADMIN' | 'MEMBER' | 'VIEWER'
}

export interface Workspace {
  id: string
  name: string
  type: 'Personal' | 'Team' | 'Competitor'
}

export interface NavigationItem {
  id: string
  label: string
  href: string
  icon: React.ComponentType<{ className?: string }>
  isActive?: boolean
}
