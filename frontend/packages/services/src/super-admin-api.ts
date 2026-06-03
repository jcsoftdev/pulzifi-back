import { getHttpClient } from '@workspace/shared-http'

export interface AdminUser {
  id: string
  email: string
  firstName: string
  lastName: string
  isSuperAdmin: boolean
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
    const { search = '', page = 1, pageSize = 5 } = params
    const query = new URLSearchParams({
      search,
      page: String(page),
      pageSize: String(pageSize),
    }).toString()
    const response = await http.get<ListUsersResponse>(`/api/v1/auth/admin/users?${query}`)
    return response
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
