import { getHttpClient } from '@workspace/shared-http'

// Internal: Backend response types (snake_case from Go)
interface ReportBackendDto {
  id: string
  page_id: string
  title: string
  report_date: string
  content: Record<string, unknown>
  pdf_url?: string
  created_by: string
  created_at: string
}

interface ListReportsBackendResponse {
  data: ReportBackendDto[]
  count: number
}

// Exported: Frontend types (camelCase)
export interface Report {
  id: string
  pageId: string
  title: string
  reportDate: string
  content: Record<string, unknown>
  pdfUrl?: string
  createdBy: string
  createdAt: string
}

export interface CreateReportDto {
  pageId: string
  title: string
  reportDate: string
  content?: Record<string, unknown>
}

// Helper: Transform backend to frontend format
function transformReport(backend: ReportBackendDto): Report {
  return {
    id: backend.id,
    pageId: backend.page_id,
    title: backend.title,
    reportDate: backend.report_date,
    content: backend.content ?? {},
    pdfUrl: backend.pdf_url,
    createdBy: backend.created_by,
    createdAt: backend.created_at,
  }
}

export const ReportApi = {
  async listReports(pageId?: string): Promise<{ data: Report[]; count: number }> {
    const http = await getHttpClient()
    const query = pageId ? `?page_id=${pageId}` : ''
    const response = await http.get<ListReportsBackendResponse>(`/api/v1/reports${query}`)
    return {
      data: (response.data ?? []).map(transformReport),
      count: response.count ?? 0,
    }
  },

  async getReport(id: string): Promise<Report> {
    const http = await getHttpClient()
    const response = await http.get<ReportBackendDto>(`/api/v1/reports/${id}`)
    return transformReport(response)
  },

  async createReport(data: CreateReportDto): Promise<Report> {
    const http = await getHttpClient()
    const payload = {
      page_id: data.pageId,
      title: data.title,
      report_date: data.reportDate,
      content: data.content ?? {},
    }
    const response = await http.post<ReportBackendDto>('/api/v1/reports', payload)
    return transformReport(response)
  },
}
