import { getHttpClient, type HttpError } from '@workspace/shared-http'

// Internal: Backend response types (snake_case from Go)
interface PageBackendDto {
  id: string
  workspace_id: string
  name: string
  url: string
  thumbnail_url?: string
  last_checked_at?: string
  last_change_detected_at?: string
  check_count: number
  tags: string[]
  check_frequency: string
  detected_changes: number
  created_by: string
  created_at: string
  updated_at: string
}

interface ListPagesBackendResponse {
  pages: PageBackendDto[]
}

// Exported: Frontend types (camelCase)
export interface Page {
  id: string
  workspaceId: string
  name: string
  url: string
  thumbnailUrl?: string
  lastCheckedAt?: string
  lastChangeDetectedAt?: string
  checkCount: number
  tags: string[]
  checkFrequency: string
  detectedChanges: number
  createdBy: string
  createdAt: string
  updatedAt: string
}

export interface CreatePageDto {
  workspaceId: string
  name: string
  url: string
  tags?: string[]
}

export interface ListPagesParams {
  workspaceId?: string
  limit?: number
}

export interface Check {
  id: string
  pageId: string
  status: string
  screenshotUrl: string
  changeDetected: boolean
  errorMessage?: string
  checkedAt: string
}

interface MonitoringConfigBackendDto {
  id: string
  page_id: string
  check_frequency: string
  schedule_type: string
  timezone: string
  block_ads_cookies: boolean
  created_at: string
  updated_at: string
}

export interface MonitoringConfig {
  id: string
  pageId: string
  checkFrequency: string
  scheduleType: string
  timezone: string
  blockAdsCookies: boolean
  createdAt: string
  updatedAt: string
}

function transformMonitoringConfig(backend: MonitoringConfigBackendDto): MonitoringConfig {
  return {
    id: backend.id,
    pageId: backend.page_id,
    checkFrequency: backend.check_frequency,
    scheduleType: backend.schedule_type,
    timezone: backend.timezone,
    blockAdsCookies: backend.block_ads_cookies,
    createdAt: backend.created_at,
    updatedAt: backend.updated_at,
  }
}

interface InsightBackendDto {
  id: string
  page_id: string
  check_id: string
  insight_type: string
  title: string
  content: string
  metadata: any
  created_at: string
}

export interface Insight {
  id: string
  pageId: string
  checkId: string
  insightType: string
  title: string
  content: string
  metadata: any
  createdAt: string
}

function transformInsight(backend: InsightBackendDto): Insight {
  return {
    id: backend.id,
    pageId: backend.page_id,
    checkId: backend.check_id,
    insightType: backend.insight_type,
    title: backend.title,
    content: backend.content,
    metadata: backend.metadata,
    createdAt: backend.created_at,
  }
}

// Helper: Transform backend to frontend format
function transformPage(backend: PageBackendDto): Page {
  return {
    id: backend.id,
    workspaceId: backend.workspace_id,
    name: backend.name,
    url: backend.url,
    thumbnailUrl: backend.thumbnail_url,
    lastCheckedAt: backend.last_checked_at,
    lastChangeDetectedAt: backend.last_change_detected_at,
    checkCount: backend.check_count,
    tags: backend.tags || [],
    checkFrequency: backend.check_frequency,
    detectedChanges: backend.detected_changes,
    createdBy: backend.created_by,
    createdAt: backend.created_at,
    updatedAt: backend.updated_at,
  }
}

export const PageApi = {
  async listPages(params?: ListPagesParams): Promise<Page[]> {
    const http = await getHttpClient()
    const queryParams = new URLSearchParams()
    if (params?.workspaceId) queryParams.append('workspace_id', params.workspaceId)
    if (params?.limit) queryParams.append('limit', params.limit.toString())

    const query = queryParams.toString() ? `?${queryParams.toString()}` : ''
    try {
      const response = await http.get<ListPagesBackendResponse>(`/api/v1/pages${query}`)
      return response.pages.map(transformPage)
    } catch (error) {
      const err = error as any
      const context = {
        endpoint: '/api/v1/pages',
        params,
        workspaceId: params?.workspaceId,
      }

      // Provide better error message for debugging
      if (err.name === 'HttpError') {
        const httpError = err as HttpError
        console.error('[PageAPI] Failed to list pages', {
          status: httpError.status,
          message: httpError.message,
          path: httpError.path,
          ...context,
        })
      } else {
        console.error('[PageAPI] Unexpected error listing pages', error, context)
      }

      throw error
    }
  },

  async getPage(id: string): Promise<Page> {
    const http = await getHttpClient()
    const response = await http.get<PageBackendDto>(`/api/v1/pages/${id}`)
    return transformPage(response)
  },

  async createPage(data: CreatePageDto): Promise<Page> {
    const http = await getHttpClient()
    const payload = {
      workspace_id: data.workspaceId,
      name: data.name,
      url: data.url,
      tags: data.tags,
    }
    const response = await http.post<PageBackendDto>('/api/v1/pages', payload)
    return transformPage(response)
  },

  async updatePage(id: string, data: Partial<CreatePageDto>): Promise<Page> {
    const http = await getHttpClient()
    const payload: Record<string, unknown> = {}
    if (data.name) payload.name = data.name
    if (data.url) payload.url = data.url
    if (data.tags) payload.tags = data.tags

    const response = await http.put<PageBackendDto>(`/api/v1/pages/${id}`, payload)
    return transformPage(response)
  },

  async deletePage(id: string): Promise<void> {
    const http = await getHttpClient()
    await http.delete(`/api/v1/pages/${id}`)
  },

  async getMonitoringConfig(pageId: string): Promise<MonitoringConfig | null> {
    const http = await getHttpClient()
    try {
      const response = await http.get<MonitoringConfigBackendDto>(
        `/api/v1/monitoring/configs/${pageId}`
      )
      return transformMonitoringConfig(response)
    } catch (error) {
      // Return null if not found or 404
      return null
    }
  },

  async updateMonitoringConfig(
    pageId: string,
    data: Partial<MonitoringConfig>
  ): Promise<MonitoringConfig> {
    const http = await getHttpClient()
    const payload: any = {}
    if (data.checkFrequency) payload.check_frequency = data.checkFrequency
    if (data.scheduleType) payload.schedule_type = data.scheduleType
    if (data.timezone) payload.timezone = data.timezone
    if (data.blockAdsCookies !== undefined) payload.block_ads_cookies = data.blockAdsCookies

    const response = await http.put<MonitoringConfigBackendDto>(
      `/api/v1/monitoring/configs/${pageId}`,
      payload
    )
    return transformMonitoringConfig(response)
  },

  async listChecks(pageId: string): Promise<Check[]> {
    const http = await getHttpClient()
    const response = await http.get<{
      checks: any[]
    }>(`/api/v1/monitoring/checks/page/${pageId}`)
    return response.checks.map((c) => ({
      id: c.id,
      pageId: c.page_id,
      status: c.status,
      screenshotUrl: c.screenshot_url,
      changeDetected: c.change_detected,
      errorMessage: c.error_message,
      checkedAt: c.checked_at,
    }))
  },

  async listInsights(pageId: string): Promise<Insight[]> {
    const http = await getHttpClient()
    const response = await http.get<{
      insights: InsightBackendDto[]
    }>(`/api/v1/insights?page_id=${pageId}`)
    return response.insights.map(transformInsight)
  },
}
