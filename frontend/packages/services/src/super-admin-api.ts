import { getHttpClient } from '@workspace/shared-http'

// Internal: Backend response types for platform invitations (snake_case from Go)
interface PlatformInvitationBackendDto {
  id: string
  email: string
  status: string
  invited_by: string
  expires_at: string
  email_sent_at?: string | null
  email_error?: string | null
  resent_count: number
  created_at: string
}

interface ListInvitationsBackendResponse {
  invitations: PlatformInvitationBackendDto[]
  total: number
}

interface InviteToPlatformBackendResponse {
  id: string
  email: string
  expires_at: string
  email_delivery: 'sent' | 'failed'
}

export type PlatformInvitationStatus = 'pending' | 'accepted' | 'revoked' | 'expired'

export interface PlatformInvitation {
  id: string
  email: string
  status: PlatformInvitationStatus
  invitedBy: string
  expiresAt: string
  emailSentAt: string | null
  emailError: string | null
  resentCount: number
  createdAt: string
}

export interface InviteToPlatformResponse {
  id: string
  email: string
  expiresAt: string
  emailDelivery: 'sent' | 'failed'
}

function transformInvitation(backend: PlatformInvitationBackendDto): PlatformInvitation {
  return {
    id: backend.id,
    email: backend.email,
    status: backend.status as PlatformInvitationStatus,
    invitedBy: backend.invited_by,
    expiresAt: backend.expires_at,
    emailSentAt: backend.email_sent_at ?? null,
    emailError: backend.email_error ?? null,
    resentCount: backend.resent_count,
    createdAt: backend.created_at,
  }
}

function transformInviteResponse(
  backend: InviteToPlatformBackendResponse
): InviteToPlatformResponse {
  return {
    id: backend.id,
    email: backend.email,
    expiresAt: backend.expires_at,
    emailDelivery: backend.email_delivery,
  }
}

export interface AdminPlan {
  id: string
  code: string
  name: string
  description: string
  checks_allowed_monthly: number
  is_active: boolean
}

export interface AdminOrganizationPlan {
  id: string
  name: string
  subdomain: string
  schema_name: string
  plan_code: string
  plan_name: string
  checks_allowed_monthly: number
}

export interface PendingUser {
  request_id: string
  user_id: string
  email: string
  first_name: string
  last_name: string
  organization_name: string
  organization_subdomain: string
  created_at: string
}

export const SuperAdminApi = {
  async listPlans(): Promise<AdminPlan[]> {
    const http = await getHttpClient()
    const response = await http.get<{
      plans: AdminPlan[]
    }>('/api/v1/usage/admin/plans')
    return response.plans || []
  },

  async listOrganizations(): Promise<AdminOrganizationPlan[]> {
    const http = await getHttpClient()
    const response = await http.get<{
      organizations: AdminOrganizationPlan[]
    }>('/api/v1/usage/admin/organizations')
    return response.organizations || []
  },

  async assignPlan(organizationId: string, planCode: string): Promise<void> {
    const http = await getHttpClient()
    await http.put(`/api/v1/usage/admin/organizations/${organizationId}/plan`, {
      plan_code: planCode,
    })
  },

  async giftMonth(organizationId: string): Promise<{
    gifted_period: {
      period_start: string
      period_end: string
      checks_allowed: number
    }
  }> {
    const http = await getHttpClient()
    return http.post(`/api/v1/usage/admin/organizations/${organizationId}/gift-month`, {})
  },

  async listPendingUsers(): Promise<PendingUser[]> {
    const http = await getHttpClient()
    const response = await http.get<{
      pending_users: PendingUser[]
    }>('/api/v1/admin/users/pending')
    return response.pending_users || []
  },

  async approveUser(requestId: string): Promise<void> {
    const http = await getHttpClient()
    await http.put(`/api/v1/admin/users/${requestId}/approve`, {})
  },

  async rejectUser(requestId: string): Promise<void> {
    const http = await getHttpClient()
    await http.put(`/api/v1/admin/users/${requestId}/reject`, {})
  },

  async inviteToPlatform(payload: { email: string }): Promise<InviteToPlatformResponse> {
    const http = await getHttpClient()
    const response = await http.post<InviteToPlatformBackendResponse>('/api/v1/admin/invitations', {
      email: payload.email,
    })
    return transformInviteResponse(response)
  },

  async listInvitations(filter?: {
    status?: PlatformInvitationStatus
    limit?: number
    offset?: number
  }): Promise<{
    invitations: PlatformInvitation[]
    total: number
  }> {
    const http = await getHttpClient()
    const params = new URLSearchParams()
    if (filter?.status) params.set('status', filter.status)
    if (filter?.limit !== undefined) params.set('limit', String(filter.limit))
    if (filter?.offset !== undefined) params.set('offset', String(filter.offset))
    const qs = params.toString()
    const url = qs ? `/api/v1/admin/invitations?${qs}` : '/api/v1/admin/invitations'
    const response = await http.get<ListInvitationsBackendResponse>(url)
    return {
      invitations: (response.invitations || []).map(transformInvitation),
      total: response.total || 0,
    }
  },

  async revokeInvitation(id: string): Promise<void> {
    const http = await getHttpClient()
    await http.delete(`/api/v1/admin/invitations/${id}`)
  },

  async resendInvitation(id: string): Promise<InviteToPlatformResponse> {
    const http = await getHttpClient()
    const response = await http.post<InviteToPlatformBackendResponse>(
      `/api/v1/admin/invitations/${id}/resend`,
      {}
    )
    return transformInviteResponse(response)
  },
}
