import { getHttpClient } from '@workspace/shared-http'

export interface AdminUser {
  id: string
  email: string
  firstName: string
  lastName: string
  isSuperAdmin: boolean
  status: string
  emailVerified: boolean
  orgCount: number
}

export interface AdminMembership {
  orgId: string
  orgName: string
  subdomain: string
  role: string
  invitationStatus: string
  isOwner: boolean
}

export interface AdminUserDetail {
  id: string
  email: string
  firstName: string
  lastName: string
  status: string
  emailVerified: boolean
  isSuperAdmin: boolean
  memberships: AdminMembership[]
}

export interface ListUsersResponse {
  users: AdminUser[]
  total: number
  page: number
  pageSize: number
}

export interface ListUsersParams {
  search?: string
  page?: number
  pageSize?: number
  orgId?: string
  status?: string
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

  // Gift = a Stripe amount_off-once coupon worth ONE MONTH of the selected
  // plan (planCode: "starter" | "pro"), credited to the org's Stripe customer
  // balance. Always one month (monthly price), never yearly. Does not change
  // the org's plan. Balance auto-applies to upcoming invoices and carries any
  // leftover forward. Requires an existing Stripe customer.
  async listUsers(params: ListUsersParams = {}): Promise<ListUsersResponse> {
    const http = await getHttpClient()
    const { search = '', page = 1, pageSize = 5, orgId = '', status = '' } = params
    const query = new URLSearchParams({
      search,
      page: String(page),
      pageSize: String(pageSize),
      orgId,
      status,
    }).toString()
    return http.get<ListUsersResponse>(`/api/v1/auth/admin/users?${query}`)
  },

  async getUser(userId: string): Promise<AdminUserDetail> {
    const http = await getHttpClient()
    return http.get<AdminUserDetail>(`/api/v1/auth/admin/users/${userId}`)
  },

  async setUserStatus(userId: string, status: 'approved' | 'suspended'): Promise<void> {
    const http = await getHttpClient()
    await http.patch(`/api/v1/auth/admin/users/${userId}/status`, { status })
  },

  async setMembershipRole(
    userId: string,
    orgId: string,
    role: 'OWNER' | 'ADMIN' | 'MEMBER'
  ): Promise<void> {
    const http = await getHttpClient()
    await http.patch(`/api/v1/auth/admin/users/${userId}/memberships/${orgId}`, { role })
  },

  async removeMembership(userId: string, orgId: string): Promise<void> {
    const http = await getHttpClient()
    await http.delete(`/api/v1/auth/admin/users/${userId}/memberships/${orgId}`)
  },

  async promoteUser(userId: string): Promise<AdminUser> {
    const http = await getHttpClient()
    return http.post<AdminUser>(`/api/v1/auth/admin/users/${userId}/promote`, {})
  },

  async giftMonth(
    organizationId: string,
    planCode: string
  ): Promise<{
    org_id: string
    customer_id: string
    gift_plan_code: string
    /** "plan_gift" (used the gifted higher plan now) | "balance_credit" (banked). */
    mode: 'plan_gift' | 'balance_credit'
    amount_cents: number
    currency: string
    /** Unix seconds the plan gift ends (0 for balance gifts). */
    revert_at: number
    message: string
  }> {
    const http = await getHttpClient()
    return http.post('/api/v1/billing/admin/gift', {
      org_id: organizationId,
      plan_code: planCode,
    })
  },
}
