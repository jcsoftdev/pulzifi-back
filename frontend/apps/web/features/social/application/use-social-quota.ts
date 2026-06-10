'use client'

import { SocialApi } from '@workspace/services/social-api'
import { useCallback, useState } from 'react'
import type { SocialQuotaStatus } from '../domain/types'

export function useSocialQuota(workspaceId: string) {
  const [quota, setQuota] = useState<SocialQuotaStatus | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const fetchQuota = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const result = await SocialApi.getQuotaStatus(workspaceId)
      setQuota(result)
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to load quota'))
    } finally {
      setIsLoading(false)
    }
  }, [workspaceId])

  /**
   * Trigger a manual check and immediately refetch quota so the badge updates.
   * Returns the resulting snapshot on success.
   */
  const triggerCheckAndRefresh = useCallback(
    async (profileId: string) => {
      const snapshot = await SocialApi.triggerCheck(profileId)
      // Refetch quota after the manual check consumes one unit (REQ-QUOTA-FE-03)
      await fetchQuota()
      return snapshot
    },
    [fetchQuota]
  )

  return {
    quota,
    isLoading,
    error,
    fetchQuota,
    triggerCheckAndRefresh,
  }
}
