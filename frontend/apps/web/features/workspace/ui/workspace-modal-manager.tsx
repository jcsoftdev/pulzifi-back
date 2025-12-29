'use client'

import { useState } from 'react'
import { useWorkspaces } from '@/features/workspace/application/hooks/use-workspaces'
import { CreateWorkspaceDialog } from './create-workspace-dialog'
import type { Workspace } from '@/features/workspace/domain/types'

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
      }
    } catch (err) {
      console.error('Failed to create workspace:', err)
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
      }
    } catch (err) {
      console.error('Failed to update workspace:', err)
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
