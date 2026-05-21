'use client'

import { UsageApi } from '@workspace/services'
import { useCallback, useEffect, useState } from 'react'
import type { TrialStatus } from '../domain/trial-status'

interface UseTrialStatusReturn {
  trial: TrialStatus | null
  isLoading: boolean
  error: string | null
  refresh: () => Promise<void>
}

/**
 * Client hook that fetches GET /api/v1/usage/trial-status for the
 * authenticated tenant. Returns null on error so the layout can fall back
 * to the regular shell without the trial banner.
 */
export function useTrialStatus(): UseTrialStatusReturn {
  const [trial, setTrial] = useState<TrialStatus | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetch = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await UsageApi.getTrialStatus()
      setTrial(data as TrialStatus)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load trial status.')
      setTrial(null)
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    fetch()
  }, [
    fetch,
  ])

  return {
    trial,
    isLoading,
    error,
    refresh: fetch,
  }
}
