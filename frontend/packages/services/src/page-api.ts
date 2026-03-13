import { getHttpClient, HttpError, getTenantFromWindow, refreshAndRetry } from '@workspace/shared-http'

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
  checkFrequency?: string
  blockAdsCookies?: boolean
  scheduleType?: string
  enabledInsightTypes?: string[]
  enabledAlertConditions?: string[]
  customAlertCondition?: string
  selectorType?: 'full_page' | 'element' | 'sections'
  cssSelector?: string
  xpathSelector?: string
  selectorOffsets?: SelectorOffsets
  sections?: SaveSectionDto[]
}

export interface ListPagesParams {
  workspaceId?: string
  limit?: number
}

export interface Check {
  id: string
  pageId: string
  sectionId?: string
  parentCheckId?: string
  status: string
  screenshotUrl: string
  htmlSnapshotUrl: string
  changeDetected: boolean
  changeType: string
  errorMessage?: string
  extractorFailed: boolean
  checkedAt: string
  sections?: Check[]
}

export interface CheckBackendDto {
  id: string
  page_id: string
  section_id?: string
  parent_check_id?: string
  status: string
  screenshot_url: string
  html_snapshot_url: string
  change_detected: boolean
  change_type: string
  error_message?: string
  checked_at: string
  sections?: CheckBackendDto[]
}

export function mapBackendCheck(c: CheckBackendDto): Check {
  return {
    id: c.id,
    pageId: c.page_id,
    sectionId: c.section_id,
    parentCheckId: c.parent_check_id,
    status: c.status,
    screenshotUrl: c.screenshot_url,
    htmlSnapshotUrl: c.html_snapshot_url,
    changeDetected: c.change_detected,
    changeType: c.change_type,
    errorMessage: c.error_message,
    extractorFailed:
      (c.status === 'error' || c.status === 'failed') &&
      typeof c.error_message === 'string' &&
      c.error_message.toLowerCase().includes('extractor'),
    checkedAt: c.checked_at,
    sections: c.sections?.map(mapBackendCheck),
  }
}

// Monitored Section types
export interface SectionRect {
  x: number
  y: number
  w: number
  h: number
}

export interface MonitoredSection {
  id: string
  pageId: string
  name: string
  cssSelector: string
  xpathSelector: string
  selectorOffsets?: SelectorOffsets
  rect?: SectionRect
  viewportWidth?: number
  sortOrder: number
  createdAt: string
  updatedAt: string
}

interface MonitoredSectionBackendDto {
  id: string
  page_id: string
  name: string
  css_selector: string
  xpath_selector: string
  selector_offsets?: SelectorOffsets
  rect?: SectionRect
  viewport_width?: number
  sort_order: number
  created_at: string
  updated_at: string
}

function transformSection(backend: MonitoredSectionBackendDto): MonitoredSection {
  return {
    id: backend.id,
    pageId: backend.page_id,
    name: backend.name,
    cssSelector: backend.css_selector,
    xpathSelector: backend.xpath_selector,
    selectorOffsets: backend.selector_offsets,
    rect: backend.rect,
    viewportWidth: backend.viewport_width,
    sortOrder: backend.sort_order,
    createdAt: backend.created_at,
    updatedAt: backend.updated_at,
  }
}

export interface SaveSectionDto {
  name: string
  cssSelector: string
  xpathSelector?: string
  selectorOffsets?: SelectorOffsets
  rect?: SectionRect
  viewportWidth?: number
  sortOrder?: number
}

export interface SelectorOffsets {
  top: number
  right: number
  bottom: number
  left: number
}

interface MonitoringConfigBackendDto {
  id: string
  page_id: string
  check_frequency: string
  schedule_type: string
  timezone: string
  block_ads_cookies: boolean
  enabled_insight_types: string[]
  enabled_alert_conditions: string[]
  custom_alert_condition: string
  selector_type: string
  css_selector: string
  xpath_selector: string
  selector_offsets?: SelectorOffsets
  created_at: string
  updated_at: string
  quota_exceeded?: boolean
}

export interface MonitoringConfig {
  id: string
  pageId: string
  checkFrequency: string
  scheduleType: string
  timezone: string
  blockAdsCookies: boolean
  enabledInsightTypes: string[]
  enabledAlertConditions: string[]
  customAlertCondition: string
  selectorType: string
  cssSelector: string
  xpathSelector: string
  selectorOffsets?: SelectorOffsets
  createdAt: string
  updatedAt: string
  quotaExceeded?: boolean
}

function transformMonitoringConfig(backend: MonitoringConfigBackendDto): MonitoringConfig {
  return {
    id: backend.id,
    pageId: backend.page_id,
    checkFrequency: backend.check_frequency,
    scheduleType: backend.schedule_type,
    timezone: backend.timezone,
    blockAdsCookies: backend.block_ads_cookies,
    enabledInsightTypes: backend.enabled_insight_types ?? ['marketing', 'market_analysis'],
    enabledAlertConditions: backend.enabled_alert_conditions ?? ['any_changes'],
    customAlertCondition: backend.custom_alert_condition ?? '',
    selectorType: backend.selector_type ?? 'full_page',
    cssSelector: backend.css_selector ?? '',
    xpathSelector: backend.xpath_selector ?? '',
    selectorOffsets: backend.selector_offsets,
    createdAt: backend.created_at,
    updatedAt: backend.updated_at,
    quotaExceeded: backend.quota_exceeded,
  }
}

export interface PreviewElement {
  selector: string
  xpath: string
  tag: string
  rect: { x: number; y: number; w: number; h: number }
  text_preview: string
  semantic_role: string
}

export interface PagePreviewResult {
  screenshot_base64: string
  viewport: { width: number; height: number }
  page_height: number
  elements: PreviewElement[]
}

export interface PreviewProgress {
  message: string
  step: number
  total_steps: number
}

interface InsightBackendDto {
  id: string
  page_id: string
  check_id: string
  insight_type: string
  title: string
  content: string
  metadata: Record<string, unknown>
  created_at: string
}

export interface Insight {
  id: string
  pageId: string
  checkId: string
  insightType: string
  title: string
  content: string
  metadata: Record<string, unknown>
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

/** Maps raw scraper/backend error messages to user-friendly text. */
function toFriendlyPreviewError(raw: string): string {
  const lower = raw.toLowerCase()
  if (lower.includes('navigation') || lower.includes('net::err_')) {
    return 'Could not reach this page. Please check the URL and try again.'
  }
  if (lower.includes('timeout') || lower.includes('timed out')) {
    return 'The page took too long to load. Please try again — it may work on a second attempt.'
  }
  if (lower.includes('browser') || lower.includes('crashed')) {
    return 'Our preview service is temporarily unavailable. Please try again in a moment.'
  }
  if (lower.includes('failed to parse')) {
    return 'We received an unexpected response from this page. Please try again.'
  }
  // Fall through: return the original message for unexpected errors
  return raw
}

export const PageApi = {
  async previewPage(url: string, blockAdsCookies = false): Promise<PagePreviewResult> {
    const http = await getHttpClient()
    return http.post<PagePreviewResult>('/api/v1/pages/preview', {
      url,
      block_ads_cookies: blockAdsCookies,
    }, { timeout: 150_000 })
  },

  /**
   * Preview a page with real-time progress streaming via SSE.
   * Uses fetch + ReadableStream to parse SSE events from a POST request.
   *
   * @param onProgress Called for each progress event with message, step, total_steps
   * @param signal Optional AbortSignal for cancellation
   * @returns The final PagePreviewResult
   */
  async previewPageWithProgress(
    url: string,
    blockAdsCookies = false,
    onProgress?: (progress: PreviewProgress) => void,
    signal?: AbortSignal,
  ): Promise<PagePreviewResult> {
    const origin = globalThis.window?.location?.origin ?? ''
    const headers: Record<string, string> = { 'Content-Type': 'application/json' }
    const tenant = getTenantFromWindow()
    if (tenant) {
      headers['X-Tenant'] = tenant
    }

    const doFetch = () =>
      fetch(`${origin}/api/v1/pages/preview`, {
        method: 'POST',
        headers,
        credentials: 'same-origin',
        body: JSON.stringify({ url, block_ads_cookies: blockAdsCookies }),
        signal,
      })

    let resp = await doFetch()

    // If unauthorized, refresh the token and retry once
    if (resp.status === 401) {
      const refreshed = await refreshAndRetry()
      if (refreshed) {
        resp = await doFetch()
      }
    }

    if (!resp.ok) {
      throw new Error(
        resp.status === 502 || resp.status === 504
          ? 'Could not reach this page. Please check the URL and try again.'
          : resp.status >= 500
            ? 'Our preview service is temporarily unavailable. Please try again in a moment.'
            : `Preview failed (${resp.status}). Please try again.`,
      )
    }

    const reader = resp.body?.getReader()
    if (!reader) {
      throw new Error('No response body')
    }

    const decoder = new TextDecoder()
    let buffer = ''
    let result: PagePreviewResult | null = null
    let error: string | null = null

    const parseFrame = (frame: string) => {
      let eventType = 'message'
      let data = ''

      for (const line of frame.split('\n')) {
        if (line.startsWith('event: ')) {
          eventType = line.slice(7)
        } else if (line.startsWith('data: ')) {
          data = line.slice(6)
        }
      }

      if (!data) return

      try {
        const parsed = JSON.parse(data)

        if (eventType === 'progress' && onProgress) {
          onProgress(parsed as PreviewProgress)
        } else if (eventType === 'result') {
          result = parsed as PagePreviewResult
        } else if (eventType === 'error') {
          error = parsed.error ?? 'Unknown preview error'
        }
      } catch (e) {
        console.error('[PageAPI] Failed to parse SSE data', {
          eventType,
          dataLength: data.length,
          dataPreview: data.slice(0, 200),
          error: e,
        })
        error = `Failed to parse preview response (${data.length} bytes)`
      }
    }

    try {
      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        // Normalize \r\n to \n — sse_starlette (Python) uses \r\n as the
        // default SSE separator, so \r\n\r\n would never match a \n\n search.
        buffer += decoder.decode(value, { stream: true }).replaceAll('\r\n', '\n')

        // Parse complete SSE frames (delimited by double newline)
        for (
          let boundary = buffer.indexOf('\n\n');
          boundary !== -1;
          boundary = buffer.indexOf('\n\n')
        ) {
          const frame = buffer.slice(0, boundary)
          buffer = buffer.slice(boundary + 2)
          parseFrame(frame)
        }
      }

      // Flush remaining bytes from the decoder
      buffer += decoder.decode()

      // Process any leftover frame in the buffer (stream may close without
      // a trailing \n\n, e.g. when the upstream connection is reset).
      const remaining = buffer.trim()
      if (remaining && !result && !error) {
        parseFrame(remaining)
      }
    } finally {
      reader.releaseLock()
    }

    if (error) {
      throw new Error(toFriendlyPreviewError(error))
    }
    if (!result) {
      throw new Error(
        'The page took too long to load or the connection was interrupted. Please try again.',
      )
    }

    return result
  },

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
      const context = {
        endpoint: '/api/v1/pages',
        params,
        workspaceId: params?.workspaceId,
      }

      // Provide better error message for debugging
      if (error instanceof HttpError) {
        console.error('[PageAPI] Failed to list pages', {
          status: error.status,
          message: error.message,
          path: error.path,
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
    const page = transformPage(response)

    // Create monitoring config via upsert with the selected settings
    const configPayload: Record<string, unknown> = {
      check_frequency: data.checkFrequency ?? 'Off',
      schedule_type: data.scheduleType ?? 'all_time',
      block_ads_cookies: data.blockAdsCookies ?? true,
      enabled_insight_types: data.enabledInsightTypes ?? ['marketing', 'market_analysis'],
      enabled_alert_conditions: data.enabledAlertConditions ?? ['any_changes'],
      custom_alert_condition: data.customAlertCondition ?? '',
      selector_type: data.selectorType ?? 'full_page',
      css_selector: data.cssSelector ?? '',
      xpath_selector: data.xpathSelector ?? '',
      selector_offsets: data.selectorOffsets ?? { top: 0, right: 0, bottom: 0, left: 0 },
    }
    try {
      await http.put(`/api/v1/monitoring/configs/${page.id}`, configPayload)
      // Patch the page object with the frequency that was actually saved,
      // because POST /api/v1/pages always returns check_frequency = "Off".
      page.checkFrequency = data.checkFrequency ?? 'Off'
    } catch (err) {
      console.error('[PageAPI] Failed to create monitoring config for page', page.id, err)
    }

    // Save sections if provided (multi-section monitoring)
    if (data.sections && data.sections.length > 0) {
      try {
        const sectionPayload = data.sections.map((s, i) => ({
          name: s.name,
          css_selector: s.cssSelector,
          xpath_selector: s.xpathSelector ?? '',
          selector_offsets: s.selectorOffsets ?? { top: 0, right: 0, bottom: 0, left: 0 },
          rect: s.rect,
          viewport_width: s.viewportWidth,
          sort_order: s.sortOrder ?? i,
        }))
        await http.post(`/api/v1/monitoring/sections/page/${page.id}`, {
          sections: sectionPayload,
        })
      } catch (err) {
        console.error('[PageAPI] Failed to save sections for page', page.id, err)
      }
    }

    return page
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

  async bulkDeletePages(ids: string[]): Promise<void> {
    const http = await getHttpClient()
    await http.post('/api/v1/pages/bulk-delete', { ids })
  },

  async bulkUpdateFrequency(pageIds: string[], checkFrequency: string): Promise<void> {
    const http = await getHttpClient()
    await http.put('/api/v1/monitoring/configs/bulk', {
      page_ids: pageIds,
      check_frequency: checkFrequency,
    })
  },

  async getMonitoringConfig(pageId: string): Promise<MonitoringConfig | null> {
    const http = await getHttpClient()
    try {
      const response = await http.get<MonitoringConfigBackendDto>(
        `/api/v1/monitoring/configs/${pageId}`
      )
      return transformMonitoringConfig(response)
    } catch {
      // Return null if not found or 404
      return null
    }
  },

  async updateMonitoringConfig(
    pageId: string,
    data: Partial<MonitoringConfig>
  ): Promise<MonitoringConfig> {
    const http = await getHttpClient()
    const payload: Record<string, unknown> = {}
    if (data.checkFrequency) payload.check_frequency = data.checkFrequency
    if (data.scheduleType) payload.schedule_type = data.scheduleType
    if (data.timezone) payload.timezone = data.timezone
    if (data.blockAdsCookies !== undefined) payload.block_ads_cookies = data.blockAdsCookies
    if (data.enabledInsightTypes !== undefined) payload.enabled_insight_types = data.enabledInsightTypes
    if (data.enabledAlertConditions !== undefined) payload.enabled_alert_conditions = data.enabledAlertConditions
    if (data.customAlertCondition !== undefined) payload.custom_alert_condition = data.customAlertCondition
    if (data.selectorType !== undefined) payload.selector_type = data.selectorType
    if (data.cssSelector !== undefined) payload.css_selector = data.cssSelector
    if (data.xpathSelector !== undefined) payload.xpath_selector = data.xpathSelector
    if (data.selectorOffsets !== undefined) payload.selector_offsets = data.selectorOffsets

    const response = await http.put<MonitoringConfigBackendDto>(
      `/api/v1/monitoring/configs/${pageId}`,
      payload
    )
    return transformMonitoringConfig(response)
  },

  async triggerCheck(pageId: string): Promise<void> {
    const http = await getHttpClient()
    await http.post(`/api/v1/monitoring/checks/page/${pageId}/run`, {})
  },

  async listChecks(pageId: string, sectionId?: string): Promise<Check[]> {
    const http = await getHttpClient()
    const query = sectionId ? `?section_id=${sectionId}` : ''
    const response = await http.get<{
      checks: CheckBackendDto[]
    }>(`/api/v1/monitoring/checks/page/${pageId}${query}`)
    return response.checks.map(mapBackendCheck)
  },

  async listSections(pageId: string): Promise<MonitoredSection[]> {
    const http = await getHttpClient()
    const response = await http.get<{
      sections: MonitoredSectionBackendDto[]
    }>(`/api/v1/monitoring/sections/page/${pageId}`)
    return response.sections.map(transformSection)
  },

  async saveSections(pageId: string, sections: SaveSectionDto[]): Promise<MonitoredSection[]> {
    const http = await getHttpClient()
    const payload = sections.map((s, i) => ({
      name: s.name,
      css_selector: s.cssSelector,
      xpath_selector: s.xpathSelector ?? '',
      selector_offsets: s.selectorOffsets ?? { top: 0, right: 0, bottom: 0, left: 0 },
      rect: s.rect,
      viewport_width: s.viewportWidth,
      sort_order: s.sortOrder ?? i,
    }))
    const response = await http.post<{
      sections: MonitoredSectionBackendDto[]
    }>(`/api/v1/monitoring/sections/page/${pageId}`, { sections: payload })
    return response.sections.map(transformSection)
  },

  async deleteSection(pageId: string, sectionId: string): Promise<void> {
    const http = await getHttpClient()
    await http.delete(`/api/v1/monitoring/sections/page/${pageId}/${sectionId}`)
  },

  async listInsights(pageId: string, checkId?: string): Promise<Insight[]> {
    const http = await getHttpClient()
    const query = checkId ? `check_id=${checkId}` : `page_id=${pageId}`
    const response = await http.get<{
      insights: InsightBackendDto[]
    }>(`/api/v1/insights?${query}`)
    return response.insights.map(transformInsight)
  },

  async generateInsights(pageId: string, checkId: string): Promise<void> {
    const http = await getHttpClient()
    // Returns 202 immediately — generation runs in the background.
    // Poll GET /insights?check_id=<id> to get results.
    await http.post('/api/v1/insights/generate', {
      page_id: pageId,
      check_id: checkId,
    })
  },
}
