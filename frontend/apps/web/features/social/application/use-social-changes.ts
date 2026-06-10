'use client'

import { SocialApi } from '@workspace/services/social-api'
import { useCallback, useState } from 'react'
import type { SocialChange } from '../domain/types'

export function useSocialChanges(profileId: string, initialChanges: SocialChange[] = []) {
  const [changes, setChanges] = useState<SocialChange[]>(initialChanges)

  const fetchChanges = useCallback(async () => {
    try {
      const result = await SocialApi.listChanges(profileId)
      setChanges(result)
    } catch {
      // silent — stale data stays
    }
  }, [profileId])

  return {
    changes,
    fetchChanges,
  }
}
