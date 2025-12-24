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
}

export interface UsageStats {
  workplaces: { current: number; max: number }
  pages: { current: number; max: number }
  checks: {
    today: number
    monthly: number
    maxMonthly: number
    percentage: number
  }
}

export class UsageApi {
  static async getChecksData(): Promise<ChecksData> {
    // Mocked for development: backend not ready / or prefer mocked data for now.
    // TODO: replace with real API call when backend endpoint is available.
    return {
      current: 300,
      max: 1000,
      refillDate: "Oct 20, 2025",
    }
    // Uncomment to call real backend:
    // const http = await getHttpClient()
    // return http.get<ChecksData>('/api/usage/checks')
  }

  static async getUsageStats(): Promise<UsageStats> {
    const http = await getHttpClient()
    return http.get<UsageStats>('/api/usage/stats')
  }

  static async incrementUsage(type: 'check' | 'page'): Promise<void> {
    const http = await getHttpClient()
    await http.post('/api/usage/increment', { type })
  }
}
