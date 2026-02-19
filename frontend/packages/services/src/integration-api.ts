import { getHttpClient } from '@workspace/shared-http'

// Internal: Backend response types
interface IntegrationBackendDto {
  id: string
  service_type: string
  config: Record<string, string>
  enabled: boolean
  created_at: string
}

interface ListIntegrationsBackendResponse {
  integrations: IntegrationBackendDto[]
}

// Exported: Frontend types
export interface Integration {
  id: string
  serviceType: string
  config: Record<string, string>
  enabled: boolean
  createdAt: string
}

export interface UpsertIntegrationDto {
  service_type: string
  config: Record<string, string>
}

function transformIntegration(backend: IntegrationBackendDto): Integration {
  return {
    id: backend.id,
    serviceType: backend.service_type,
    config: backend.config ?? {},
    enabled: backend.enabled,
    createdAt: backend.created_at,
  }
}

export const IntegrationApi = {
  async listIntegrations(): Promise<Integration[]> {
    const http = await getHttpClient()
    const response = await http.get<ListIntegrationsBackendResponse>('/api/v1/integrations')
    return (response.integrations ?? []).map(transformIntegration)
  },

  async upsertIntegration(data: UpsertIntegrationDto): Promise<Integration> {
    const http = await getHttpClient()
    const response = await http.post<IntegrationBackendDto>('/api/v1/integrations', data)
    return transformIntegration(response)
  },

  async deleteIntegration(id: string): Promise<void> {
    const http = await getHttpClient()
    await http.delete(`/api/v1/integrations/${id}`)
  },
}
