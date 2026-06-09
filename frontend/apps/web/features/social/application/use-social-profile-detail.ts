'use client'

import { SocialApi } from '@workspace/services/social-api'
import { useCallback, useState } from 'react'
import type { SocialProfile, SocialSnapshot } from '../domain/types'

export function useSocialProfileDetail(profileId: string) {
  const [profile, setProfile] = useState<SocialProfile | null>(null)
  const [snapshot, setSnapshot] = useState<SocialSnapshot | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const fetchProfile = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const result = await SocialApi.getProfile(profileId)
      setProfile(result.profile)
      setSnapshot(result.latestSnapshot)
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to load profile'))
    } finally {
      setIsLoading(false)
    }
  }, [profileId])

  return {
    profile,
    snapshot,
    isLoading,
    error,
    fetchProfile,
  }
}
