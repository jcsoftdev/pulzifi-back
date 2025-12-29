/**
 * Usage Service - Domain Service
 * Handles all usage-related data fetching
 * This is feature-specific and belongs to the Usage domain
 */

import { UsageApi } from '@workspace/services'
import type { ChecksData, UsageStats } from '../types'

export const UsageService = {
  /**
   * Get current checks usage data
   * Works in both server-side and client-side
   */
  async getChecksData(): Promise<ChecksData> {
    try {
      const data = await UsageApi.getChecksData()
      return data
    } catch (error) {
      console.error('Failed to fetch checks data:', error)

      // Fallback to mock data
      return {
        current: 300,
        max: 1000,
        refillDate: 'Oct 20, 2025',
      }
    }
  },

  /**
   * Get complete usage statistics
   */
  async getUsageStats(): Promise<UsageStats> {
    try {
      const data = await UsageApi.getUsageStats()
      return data
    } catch (error) {
      console.error('Failed to fetch usage stats:', error)

      return {
        workplaces: {
          current: 3,
          max: 10,
        },
        pages: {
          current: 20,
          max: 200,
        },
        checks: {
          today: 40,
          monthly: 100,
          maxMonthly: 2000,
          percentage: 5,
        },
      }
    }
  },
}
