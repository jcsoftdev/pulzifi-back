'use client'

import { SocialApi } from '@workspace/services/social-api'
import { useCallback, useState } from 'react'
import type { SocialProfile, SocialSnapshot } from '../domain/types'

export function useSocialProfileDetail(
  profileId: string,
  initialProfile: SocialProfile,
  initialSnapshot: SocialSnapshot | null,
) {
  const [profile, setProfile] = useState<SocialProfile>(initialProfile)
  const [snapshot, setSnapshot] = useState<SocialSnapshot | null>(initialSnapshot)
  const [isChecking, setIsChecking] = useState(false)
  const [checkError, setCheckError] = useState<string | null>(null)

  const runCheck = useCallback(async (onSuccess?: () => void) => {
    setIsChecking(true)
    setCheckError(null)
    try {
      await SocialApi.triggerCheck(profileId)
      const result = await SocialApi.getProfile(profileId)
      setProfile(result.profile)
      setSnapshot(result.latestSnapshot)
      onSuccess?.()
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Check failed'
      setCheckError(
        msg.toLowerCase().includes('quota') ? 'Daily check quota exceeded' : msg
      )
    } finally {
      setIsChecking(false)
    }
  }, [profileId])

  return {
    profile,
    snapshot,
    isChecking,
    checkError,
    runCheck,
  }
}
