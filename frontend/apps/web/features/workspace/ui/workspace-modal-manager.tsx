'use client'

import { useState } from 'react'
import { notification } from '@/lib/notification'
import { useWorkspaces } from '@/features/workspace/application/hooks/use-workspaces'
import type { Workspace } from '@/features/workspace/domain/types'
import { CreateWorkspaceDialog } from './create-workspace-dialog'

export interface WorkspaceModalManagerProps {
  onWorkspaceCreated?: (workspace: Workspace) => void
  onWorkspaceUpdated?: (workspace: Workspace) => void
}

export function WorkspaceModalManager({
  onWorkspaceCreated,
  onWorkspaceUpdated,
}: Readonly<WorkspaceModalManagerProps>) {
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
  const [editingWorkspace, setEditingWorkspace] = useState<Workspace | null>(null)
  const { isLoading, error, createWorkspace, updateWorkspace } = useWorkspaces()

  const handleCreateWorkspace = async (data: {
    name: string
    type: 'Personal' | 'Team' | 'Competitor'
  }) => {
    try {
      const result = await createWorkspace(data)
      if (result) {
        setIsCreateDialogOpen(false)
        onWorkspaceCreated?.(result)
        notification.success({ title: 'Workspace created', description: `"${result.name}" is ready.` })
      }
    } catch (err) {
      console.error('Failed to create workspace:', err)
      notification.error({ title: 'Failed to create workspace', description: err instanceof Error ? err.message : 'Please try again.' })
    }
  }

  const handleUpdateWorkspace = async (data: {
    name?: string
    type?: 'Personal' | 'Team' | 'Competitor'
  }) => {
    if (!editingWorkspace) return
    try {
      const result = await updateWorkspace(editingWorkspace.id, data)
      if (result) {
        setEditingWorkspace(null)
        onWorkspaceUpdated?.(result)
        notification.success({ title: 'Workspace updated', description: `"${result.name}" has been updated.` })
      }
    } catch (err) {
      console.error('Failed to update workspace:', err)
      notification.error({ title: 'Failed to update workspace', description: err instanceof Error ? err.message : 'Please try again.' })
    }
  }

  return (
    <CreateWorkspaceDialog
      open={isCreateDialogOpen}
      onOpenChange={setIsCreateDialogOpen}
      onSubmit={editingWorkspace ? handleUpdateWorkspace : handleCreateWorkspace}
      isLoading={isLoading}
      error={error}
      initialData={editingWorkspace}
      mode={editingWorkspace ? 'edit' : 'create'}
    />
  )
}
