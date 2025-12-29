import { getHttpClient } from '@workspace/shared-http'

// Internal: Backend response types (snake_case from Go)
interface OrganizationBackendDto {
  id: string
  name: string
  company?: string
  created_at: string
  updated_at?: string
}

// Exported: Frontend types (camelCase)
export interface Organization {
  id: string
  name: string
  company?: string
  createdAt: string
  updatedAt?: string
}

// Helper: Transform backend to frontend format
function transformOrganization(backend: OrganizationBackendDto): Organization {
  return {
    id: backend.id,
    name: backend.name,
    company: backend.company,
    createdAt: backend.created_at,
    updatedAt: backend.updated_at,
  }
}

export const OrganizationApi = {
  async getCurrentOrganization(): Promise<Organization> {
    const http = await getHttpClient()
    const response = await http.get<OrganizationBackendDto>('/api/v1/organization/current')
    return transformOrganization(response)
  },

  async getOrganization(id: string): Promise<Organization> {
    const http = await getHttpClient()
    const response = await http.get<OrganizationBackendDto>(`/api/v1/organizations/${id}`)
    return transformOrganization(response)
  },
}
