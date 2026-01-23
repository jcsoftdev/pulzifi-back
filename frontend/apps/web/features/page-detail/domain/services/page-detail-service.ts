import {
  PageApi,
  type Page,
  type MonitoringConfig,
  type Check,
  type Insight,
} from '@workspace/services/page-api'
import { WorkspaceApi, type Workspace } from '@workspace/services/workspace-api'

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
