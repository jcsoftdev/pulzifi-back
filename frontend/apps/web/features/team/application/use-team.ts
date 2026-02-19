'use client'

import { TeamApi } from '@workspace/services'
import { useCallback, useEffect, useState } from 'react'
import type { TeamMember } from '../domain/types'

export function useTeam() {
  const [members, setMembers] = useState<TeamMember[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchMembers = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await TeamApi.listMembers()
      setMembers(data as TeamMember[])
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to load team members'))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchMembers()
  }, [fetchMembers])

  const inviteMember = useCallback(async (email: string, role: string) => {
    const member = await TeamApi.inviteMember({ email, role })
    setMembers((prev) => [...prev, member as TeamMember])
    return member
  }, [])

  const updateMember = useCallback(async (memberId: string, role: string) => {
    await TeamApi.updateMember(memberId, { role })
    setMembers((prev) =>
      prev.map((m) => (m.id === memberId ? { ...m, role: role as TeamMember['role'] } : m))
    )
  }, [])

  const removeMember = useCallback(async (memberId: string) => {
    await TeamApi.removeMember(memberId)
    setMembers((prev) => prev.filter((m) => m.id !== memberId))
  }, [])

  return {
    members,
    loading,
    error,
    inviteMember,
    updateMember,
    removeMember,
    refresh: fetchMembers,
  }
}
