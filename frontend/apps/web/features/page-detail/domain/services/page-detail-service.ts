import {
  type Check,
  type Insight,
  type MonitoringConfig,
  type Page,
  PageApi,
} from '@workspace/services/page-api'
import { type ChecksData, UsageApi } from '@workspace/services/usage-api'
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

  async getQuotaStatus(): Promise<{ exceeded: boolean; refillDate: string }> {
    try {
      const data: ChecksData = await UsageApi.getChecksData()
      return {
        exceeded: data.current >= data.max,
        refillDate: data.refillDate,
      }
    } catch {
      return { exceeded: false, refillDate: 'N/A' }
    }
  },
}
