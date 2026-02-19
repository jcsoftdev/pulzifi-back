'use client'

import { DashboardApi, type DashboardStats } from '@workspace/services'
import { useEffect, useState } from 'react'

export interface UseDashboardStatsResult {
  stats: DashboardStats | null
  loading: boolean
  error: Error | null
}

export function useDashboardStats(): UseDashboardStatsResult {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    DashboardApi.getStats()
      .then(setStats)
      .catch((err) => setError(err instanceof Error ? err : new Error('Failed to load dashboard stats')))
      .finally(() => setLoading(false))
  }, [])

  return { stats, loading, error }
}
