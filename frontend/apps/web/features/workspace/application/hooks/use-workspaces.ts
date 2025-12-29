'use client'

import { useCallback, useState } from 'react'
import { WorkspaceApi, type CreateWorkspaceDto } from '@workspace/services'
import type { Workspace } from '../../domain/types'

export function useWorkspaces() {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const createWorkspace = useCallback(
    async (data: CreateWorkspaceDto): Promise<Workspace | null> => {
      setIsLoading(true)
      setError(null)
      try {
        const newWorkspace = await WorkspaceApi.createWorkspace(data)
        return newWorkspace
      } catch (err) {
        const error = err instanceof Error ? err : new Error('Failed to create workspace')
        setError(error)
        throw error
      } finally {
        setIsLoading(false)
      }
    },
    []
  )

  const updateWorkspace = useCallback(
    async (id: string, data: Partial<CreateWorkspaceDto>): Promise<Workspace | null> => {
      setIsLoading(true)
      setError(null)
      try {
        const updated = await WorkspaceApi.updateWorkspace(id, data)
        return updated
      } catch (err) {
        const error = err instanceof Error ? err : new Error('Failed to update workspace')
        setError(error)
        throw error
      } finally {
        setIsLoading(false)
      }
    },
    []
  )

  const deleteWorkspace = useCallback(async (id: string) => {
    setIsLoading(true)
    setError(null)
    try {
      await WorkspaceApi.deleteWorkspace(id)
      return true
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Failed to delete workspace')
      setError(error)
      throw error
    } finally {
      setIsLoading(false)
    }
  }, [])

  return {
    isLoading,
    error,
    createWorkspace,
    updateWorkspace,
    deleteWorkspace,
  }
}
