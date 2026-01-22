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
}

export interface ListPagesParams {
  workspaceId?: string
  limit?: number
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
    }
    const response = await http.post<PageBackendDto>('/api/v1/pages', payload)
    return transformPage(response)
  },

  async updatePage(id: string, data: Partial<CreatePageDto>): Promise<Page> {
    const http = await getHttpClient()
    const payload: Record<string, unknown> = {}
    if (data.name) payload.name = data.name
    if (data.url) payload.url = data.url
    
    const response = await http.put<PageBackendDto>(`/api/v1/pages/${id}`, payload)
    return transformPage(response)
  },

  async deletePage(id: string): Promise<void> {
    const http = await getHttpClient()
    await http.delete(`/api/v1/pages/${id}`)
  },

  async updateMonitoringConfig(pageId: string, frequency: string): Promise<void> {
    const http = await getHttpClient()
    await http.put(`/api/v1/monitoring/configs/${pageId}`, {
      check_frequency: frequency,
    })
  },
}
