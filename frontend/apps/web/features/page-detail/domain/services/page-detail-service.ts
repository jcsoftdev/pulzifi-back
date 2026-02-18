import {
  type Check,
  type Insight,
  type MonitoringConfig,
  type Page,
  PageApi,
} from '@workspace/services/page-api'
import { type Workspace, WorkspaceApi } from '@workspace/services/workspace-api'

export const PageDetailService = {
  async getPage(pageId: string): Promise<Page> {
    return await PageApi.getPage(pageId)
  },

  async getWorkspace(workspaceId: string): Promise<Workspace> {
    return await WorkspaceApi.getWorkspace(workspaceId)
  },

  async getMonitoringConfig(pageId: string): Promise<MonitoringConfig | null> {
    return await PageApi.getMonitoringConfig(pageId)
  },

  async listChecks(pageId: string): Promise<Check[]> {
    return await PageApi.listChecks(pageId)
  },

  async listInsights(pageId: string): Promise<Insight[]> {
    return await PageApi.listInsights(pageId)
  },
}
