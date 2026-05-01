'use client'

import {
  type InviteToPlatformResponse,
  type PlatformInvitation,
  SuperAdminApi,
} from '@workspace/services'
import { useCallback, useEffect, useState } from 'react'

export function useInvitations() {
  const [invitations, setInvitations] = useState<PlatformInvitation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    setLoading(true)
    try {
      const { invitations: list } = await SuperAdminApi.listInvitations()
      setInvitations(list)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load invitations')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    refresh()
  }, [
    refresh,
  ])

  const create = useCallback(
    async (email: string): Promise<InviteToPlatformResponse> => {
      const res = await SuperAdminApi.inviteToPlatform({
        email,
      })
      await refresh()
      return res
    },
    [
      refresh,
    ]
  )

  const revoke = useCallback(
    async (id: string) => {
      await SuperAdminApi.revokeInvitation(id)
      await refresh()
    },
    [
      refresh,
    ]
  )

  const resend = useCallback(
    async (id: string): Promise<InviteToPlatformResponse> => {
      const res = await SuperAdminApi.resendInvitation(id)
      await refresh()
      return res
    },
    [
      refresh,
    ]
  )

  return {
    invitations,
    loading,
    error,
    create,
    revoke,
    resend,
    refresh,
  }
}
