import { getHttpClient } from '@workspace/shared-http'

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

  async listPendingUsers(): Promise<PendingUser[]> {
    const http = await getHttpClient()
    const response = await http.get<{
      requests: PendingUser[]
    }>('/api/v1/admin/users/pending')
    return response.requests || []
  },

  async approveUser(requestId: string): Promise<void> {
    const http = await getHttpClient()
    await http.put(`/api/v1/admin/users/${requestId}/approve`, {})
  },

  async rejectUser(requestId: string): Promise<void> {
    const http = await getHttpClient()
    await http.put(`/api/v1/admin/users/${requestId}/reject`, {})
  },
}
