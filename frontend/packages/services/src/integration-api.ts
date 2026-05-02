import { getHttpClient } from '@workspace/shared-http'

// ---------------------------------------------------------------------------
// Enums / value types
// ---------------------------------------------------------------------------

export type IntegrationStatus = 'active' | 'disconnected' | 'expired'

export type DeliveryStatus = 'pending' | 'in_flight' | 'delivered' | 'failed' | 'dead'

export type ScopeType = 'org' | 'workspace' | 'page'

// ---------------------------------------------------------------------------
// Backend DTOs (snake_case — matches what the Go server returns)
// ---------------------------------------------------------------------------

interface IntegrationBackendDto {
  id: string
  service_type: string
  status: IntegrationStatus
  provider_meta: Record<string, unknown>
  created_at: string
  updated_at: string
}

interface TargetBackendDto {
  id: string
  name: string
  meta?: Record<string, unknown>
}

interface DestinationBackendDto {
  id: string
  integration_id?: string | null
  service_type: string
  scope_type: ScopeType
  scope_id: string
  target: Record<string, unknown>
  events: string[]
  enabled: boolean
  created_at: string
  updated_at: string
}

interface DeliveryBackendDto {
  id: string
  destination_id: string
  event_type: string
  event_payload: Record<string, unknown>
  status: DeliveryStatus
  attempts: number
  last_attempt_at?: string | null
  next_attempt_at?: string | null
  response_code?: number | null
  response_body?: string | null
  error_message?: string | null
  delivered_at?: string | null
  created_at: string
}

// ---------------------------------------------------------------------------
// Frontend types (camelCase — what the rest of the app uses)
// ---------------------------------------------------------------------------

export interface Integration {
  id: string
  serviceType: string
  status: IntegrationStatus
  providerMeta: Record<string, unknown>
  createdAt: string
  updatedAt: string
}

export interface Target {
  id: string
  name: string
  meta?: Record<string, unknown>
}

export interface Destination {
  id: string
  integrationId?: string | null
  serviceType: string
  scopeType: ScopeType
  scopeId: string
  target: Record<string, unknown>
  events: string[]
  enabled: boolean
  createdAt: string
  updatedAt: string
}

export interface Delivery {
  id: string
  destinationId: string
  eventType: string
  eventPayload: Record<string, unknown>
  status: DeliveryStatus
  attempts: number
  lastAttemptAt?: string | null
  nextAttemptAt?: string | null
  responseCode?: number | null
  responseBody?: string | null
  errorMessage?: string | null
  deliveredAt?: string | null
  createdAt: string
}

export interface CreateDestinationInput {
  integration_id?: string
  service_type: string
  scope_type: ScopeType
  scope_id: string
  target: Record<string, unknown>
  events: string[]
}

export interface UpdateDestinationInput {
  target?: Record<string, unknown>
  events?: string[]
  enabled?: boolean
}

// ---------------------------------------------------------------------------
// Transform helpers
// ---------------------------------------------------------------------------

function transformIntegration(dto: IntegrationBackendDto): Integration {
  return {
    id: dto.id,
    serviceType: dto.service_type,
    status: dto.status,
    providerMeta: dto.provider_meta ?? {},
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
  }
}

function transformTarget(dto: TargetBackendDto): Target {
  return {
    id: dto.id,
    name: dto.name,
    meta: dto.meta,
  }
}

function transformDestination(dto: DestinationBackendDto): Destination {
  return {
    id: dto.id,
    integrationId: dto.integration_id,
    serviceType: dto.service_type,
    scopeType: dto.scope_type,
    scopeId: dto.scope_id,
    target: dto.target ?? {},
    events: dto.events ?? [],
    enabled: dto.enabled,
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
  }
}

function transformDelivery(dto: DeliveryBackendDto): Delivery {
  return {
    id: dto.id,
    destinationId: dto.destination_id,
    eventType: dto.event_type,
    eventPayload: dto.event_payload ?? {},
    status: dto.status,
    attempts: dto.attempts,
    lastAttemptAt: dto.last_attempt_at,
    nextAttemptAt: dto.next_attempt_at,
    responseCode: dto.response_code,
    responseBody: dto.response_body,
    errorMessage: dto.error_message,
    deliveredAt: dto.delivered_at,
    createdAt: dto.created_at,
  }
}

// ---------------------------------------------------------------------------
// API objects
// ---------------------------------------------------------------------------

export const IntegrationsApi = {
  async list(): Promise<Integration[]> {
    const http = await getHttpClient()
    const response = await http.get<{ data: IntegrationBackendDto[] }>('/api/v1/integrations')
    return (response.data ?? []).map(transformIntegration)
  },

  async disconnect(id: string): Promise<void> {
    const http = await getHttpClient()
    await http.delete(`/api/v1/integrations/${id}`)
  },

  /** Returns the redirect URL the browser should navigate to in order to start OAuth. */
  startOAuthURL(provider: string, returnPath = '/settings/integrations'): string {
    return `/api/v1/integrations/oauth/${provider}/start?return_path=${encodeURIComponent(returnPath)}`
  },

  async listTargets(id: string): Promise<Target[]> {
    const http = await getHttpClient()
    const response = await http.get<{ targets: TargetBackendDto[] }>(
      `/api/v1/integrations/${id}/targets`
    )
    return (response.targets ?? []).map(transformTarget)
  },
}

export const DestinationsApi = {
  async list(params: {
    scope_type: ScopeType
    scope_id: string
    service_type?: string
  }): Promise<Destination[]> {
    const http = await getHttpClient()
    const qs = new URLSearchParams({
      scope_type: params.scope_type,
      scope_id: params.scope_id,
      ...(params.service_type ? { service_type: params.service_type } : {}),
    }).toString()
    const response = await http.get<{ data: DestinationBackendDto[] }>(
      `/api/v1/destinations?${qs}`
    )
    return (response.data ?? []).map(transformDestination)
  },

  async create(body: CreateDestinationInput): Promise<Destination> {
    const http = await getHttpClient()
    const response = await http.post<DestinationBackendDto>('/api/v1/destinations', body)
    return transformDestination(response)
  },

  async update(id: string, body: UpdateDestinationInput): Promise<Destination> {
    const http = await getHttpClient()
    const response = await http.patch<DestinationBackendDto>(`/api/v1/destinations/${id}`, body)
    return transformDestination(response)
  },

  async delete(id: string): Promise<void> {
    const http = await getHttpClient()
    await http.delete(`/api/v1/destinations/${id}`)
  },
}

export const DeliveriesApi = {
  async list(params: {
    destination_id: string
    status?: string
    limit?: number
    offset?: number
  }): Promise<{ data: Delivery[] }> {
    const http = await getHttpClient()
    const qs = new URLSearchParams({
      destination_id: params.destination_id,
      ...(params.status ? { status: params.status } : {}),
      ...(params.limit !== undefined ? { limit: String(params.limit) } : {}),
      ...(params.offset !== undefined ? { offset: String(params.offset) } : {}),
    }).toString()
    const response = await http.get<{ data: DeliveryBackendDto[] }>(`/api/v1/deliveries?${qs}`)
    return { data: (response.data ?? []).map(transformDelivery) }
  },

  async retry(id: string): Promise<void> {
    const http = await getHttpClient()
    await http.post(`/api/v1/deliveries/${id}/retry`, {})
  },
}
