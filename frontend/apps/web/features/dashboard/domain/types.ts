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
