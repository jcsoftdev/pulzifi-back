'use client'

import { SocialApi } from '@workspace/services/social-api'
import type { SocialSnapshotSummary } from '@workspace/services/social-api'
import { useCallback, useState } from 'react'

export function useSocialSnapshots(profileId: string, initialSnapshots: SocialSnapshotSummary[] = []) {
  const [snapshots, setSnapshots] = useState<SocialSnapshotSummary[]>(initialSnapshots)
  const [isLoading, setIsLoading] = useState(false)

  const fetchSnapshots = useCallback(async () => {
    setIsLoading(true)
    try {
      const result = await SocialApi.listSnapshots(profileId)
      setSnapshots(result)
    } catch {
      // silent
    } finally {
      setIsLoading(false)
    }
  }, [profileId])

  return { snapshots, isLoading, fetchSnapshots }
}
