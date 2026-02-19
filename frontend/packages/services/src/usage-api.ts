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
}

interface UsageQuotasResponse {
  quotas: {
    checks_used?: number
    checks_allowed?: number
    next_refill_at?: string | null
    storage_period_days?: number
  }
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

    return {
      current,
      max,
      refillDate,
      storagePeriodDays,
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
}
