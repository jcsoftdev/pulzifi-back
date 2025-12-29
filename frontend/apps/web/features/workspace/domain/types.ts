import type { Workspace as WorkspaceService } from '@workspace/services'

// Re-export from services (already in camelCase)
export type Workspace = WorkspaceService

export type WorkspaceType = 'Personal' | 'Team' | 'Competitor'
export type WorkspaceStatus = 'Active' | 'Deleted'

export interface WorkspaceFilters {
  search: string
}

export interface WorkspaceStats {
  lastCheckTime: string
}
