'use client'

import { SocialApi } from '@workspace/services/social-api'
import { useCallback, useState } from 'react'
import type { CreateSocialProfileDto, SocialProfile } from '../domain/types'

export function useSocialProfiles(workspaceId: string) {
  const [profiles, setProfiles] = useState<SocialProfile[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const fetchProfiles = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const result = await SocialApi.listProfiles(workspaceId)
      setProfiles(result)
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to load profiles'))
    } finally {
      setIsLoading(false)
    }
  }, [workspaceId])

  const createProfile = useCallback(
    async (data: CreateSocialProfileDto): Promise<SocialProfile> => {
      const created = await SocialApi.createProfile(data)
      setProfiles((prev) => [created, ...prev])
      return created
    },
    []
  )

  const deleteProfile = useCallback(async (profileId: string) => {
    await SocialApi.deleteProfile(profileId)
    setProfiles((prev) => prev.filter((p) => p.id !== profileId))
  }, [])

  return {
    profiles,
    isLoading,
    error,
    fetchProfiles,
    createProfile,
    deleteProfile,
  }
}
