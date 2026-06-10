'use client'

import { SocialApi } from '@workspace/services/social-api'
import { useCallback, useEffect, useState } from 'react'
import type { SocialChange } from '../domain/types'

export function useSocialWorkspaceChanges(workspaceId: string) {
  const [changes, setChanges] = useState<SocialChange[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const fetchChanges = useCallback(async () => {
    if (!workspaceId) return
    setIsLoading(true)
    setError(null)
    try {
      const result = await SocialApi.listWorkspaceChanges(workspaceId)
      setChanges(result)
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to load social changes'))
    } finally {
      setIsLoading(false)
    }
  }, [workspaceId])

  useEffect(() => {
    fetchChanges()
  }, [fetchChanges])

  return { changes, isLoading, error, refetch: fetchChanges }
}
