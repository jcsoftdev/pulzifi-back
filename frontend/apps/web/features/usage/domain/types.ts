/**
 * Usage Feature - Domain Types
 */

export interface ChecksData {
  current: number
  max: number
  refillDate: string
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
