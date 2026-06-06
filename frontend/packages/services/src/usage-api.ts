/**
 * Usage API Service
 * Consumes backend /api/usage/* endpoints
 * Works dynamically in both server-side and client-side
 */

import { getHttpClient } from '@workspace/shared-http'

export interface ChecksData {
  current: number
  max: number
  refillDate: string
  storagePeriodDays: number
  maxPages: number
  maxWorkspaces: number
}

interface UsageQuotasResponse {
  quotas: {
    checks_used?: number
    checks_allowed?: number
    next_refill_at?: string | null
    storage_period_days?: number
    max_pages?: number
    max_workspaces?: number
  }
}

export interface TrialStatusDto {
  is_trial: boolean
  is_expired: boolean
  needs_upgrade: boolean
  converted: boolean
  days_remaining: number
  trial_ends_at?: string
}

export interface UsageStats {
  workplaces: {
    current: number
    max: number
  }
  pages: {
    current: number
    max: number
  }
  checks: {
    today: number
    monthly: number
    maxMonthly: number
    percentage: number
  }
}

export const UsageApi = {
  async getChecksData(): Promise<ChecksData> {
    const http = await getHttpClient()
    const response = await http.get<UsageQuotasResponse>('/api/v1/usage/quotas')

    const current = response.quotas?.checks_used ?? 0
    const max = response.quotas?.checks_allowed ?? 0
    const nextRefillRaw = response.quotas?.next_refill_at

    let refillDate = 'N/A'
    if (nextRefillRaw) {
      const parsed = new Date(nextRefillRaw)
      if (!Number.isNaN(parsed.getTime())) {
        refillDate = parsed.toLocaleDateString('en-US', {
          month: 'short',
          day: '2-digit',
          year: 'numeric',
        })
      }
    }

    const storagePeriodDays = response.quotas?.storage_period_days ?? 7
    const maxPages = response.quotas?.max_pages ?? 0
    const maxWorkspaces = response.quotas?.max_workspaces ?? 0

    return {
      current,
      max,
      refillDate,
      storagePeriodDays,
      maxPages,
      maxWorkspaces,
    }
  },

  async getUsageStats(): Promise<UsageStats> {
    const http = await getHttpClient()
    return http.get<UsageStats>('/api/usage/stats')
  },

  async incrementUsage(type: 'check' | 'page'): Promise<void> {
    const http = await getHttpClient()
    await http.post('/api/usage/increment', {
      type,
    })
  },

  /**
   * Fetches the current trial status for the authenticated tenant.
   * Used by the trial banner + upgrade modal on the (main) layout.
   */
  async getTrialStatus(): Promise<TrialStatusDto> {
    const http = await getHttpClient()
    return http.get<TrialStatusDto>('/api/v1/usage/trial-status')
  },
}
