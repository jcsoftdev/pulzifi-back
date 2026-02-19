export interface DashboardStats {
  workplaces: number
  maxWorkplaces: number
  pages: number
  maxPages: number
  todayChecks: number
  monthlyChecks: number
  maxMonthlyChecks: number
  usagePercent: number
}

export interface WorkspaceItem {
  id: string
  name: string
  type: 'Personal' | 'Team' | 'Competitor'
  pageCount?: number
}

export interface WorkspaceChanges {
  workspaceName: string
  detectedChanges: number
}

export interface RecentAlert {
  checkedAt: string
  workspaceName: string
  changeType: string
  pageUrl: string
}

export interface RecentInsight {
  createdAt: string
  workspaceName: string
  pageUrl: string
  title: string
  content: string
}
