'use client'

import { DashboardApi, type DashboardStats } from '@workspace/services'
import { useCallback, useEffect, useState } from 'react'

export interface UseDashboardStatsResult {
  stats: DashboardStats | null
  loading: boolean
  error: Error | null
  refetch: () => void
}

export function useDashboardStats(): UseDashboardStatsResult {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [fetchKey, setFetchKey] = useState(0)

  useEffect(() => {
    setLoading(true)
    DashboardApi.getStats()
      .then(setStats)
      .catch((err) => setError(err instanceof Error ? err : new Error('Failed to load dashboard stats')))
      .finally(() => setLoading(false))
  }, [fetchKey])

  const refetch = useCallback(() => setFetchKey((k) => k + 1), [])

  return { stats, loading, error, refetch }
}
