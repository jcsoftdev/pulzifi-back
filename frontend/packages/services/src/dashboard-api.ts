import { getHttpClient } from '@workspace/shared-http'

// Internal: Backend response types (snake_case from Go)
interface WorkspaceChangesBackendDto {
  workspace_name: string
  detected_changes: number
}

interface RecentAlertBackendDto {
  checked_at: string
  workspace_name: string
  change_type: string
  page_url: string
}

interface RecentInsightBackendDto {
  created_at: string
  workspace_name: string
  page_url: string
  title: string
  content: string
}

interface DashboardStatsBackendDto {
  workspaces_count: number
  pages_count: number
  today_checks_count: number
  changes_per_workspace: WorkspaceChangesBackendDto[]
  recent_alerts: RecentAlertBackendDto[]
  recent_insights: RecentInsightBackendDto[]
}

// Exported: Frontend types (camelCase)
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

export interface DashboardStats {
  workspacesCount: number
  pagesCount: number
  todayChecksCount: number
  changesPerWorkspace: WorkspaceChanges[]
  recentAlerts: RecentAlert[]
  recentInsights: RecentInsight[]
}

function transformDashboardStats(backend: DashboardStatsBackendDto): DashboardStats {
  return {
    workspacesCount: backend.workspaces_count,
    pagesCount: backend.pages_count,
    todayChecksCount: backend.today_checks_count,
    changesPerWorkspace: (backend.changes_per_workspace ?? []).map((c) => ({
      workspaceName: c.workspace_name,
      detectedChanges: c.detected_changes,
    })),
    recentAlerts: (backend.recent_alerts ?? []).map((a) => ({
      checkedAt: a.checked_at,
      workspaceName: a.workspace_name,
      changeType: a.change_type,
      pageUrl: a.page_url,
    })),
    recentInsights: (backend.recent_insights ?? []).map((i) => ({
      createdAt: i.created_at,
      workspaceName: i.workspace_name,
      pageUrl: i.page_url,
      title: i.title,
      content: i.content,
    })),
  }
}

export const DashboardApi = {
  async getStats(): Promise<DashboardStats> {
    const http = await getHttpClient()
    const response = await http.get<DashboardStatsBackendDto>('/api/v1/dashboard/stats')
    return transformDashboardStats(response)
  },
}
