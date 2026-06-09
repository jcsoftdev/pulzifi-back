'use client'

import { SocialApi } from '@workspace/services/social-api'
import { useCallback, useState } from 'react'
import type { SocialChange } from '../domain/types'

export function useSocialChanges(profileId: string) {
  const [changes, setChanges] = useState<SocialChange[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const fetchChanges = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const result = await SocialApi.listChanges(profileId)
      setChanges(result)
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to load changes'))
    } finally {
      setIsLoading(false)
    }
  }, [profileId])

  return {
    changes,
    isLoading,
    error,
    fetchChanges,
  }
}
