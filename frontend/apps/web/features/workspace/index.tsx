'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { SquarePlus } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { useMemo, useState } from 'react'
import { useWorkspaces } from './application/hooks/use-workspaces'
import type { Workspace } from './domain/types'
import { CreateWorkspaceDialog } from './ui/create-workspace-dialog'
import { DeleteWorkspaceDialog } from './ui/delete-workspace-dialog'
import { EmptyState } from './ui/empty-state'
import { SearchBar } from './ui/search-bar'
import { WelcomeContainer } from './ui/welcome-container'
import { WorkspaceCard } from './ui/workspace-card'

export interface WorkspaceFeatureProps {
  initialWorkspaces?: Workspace[]
  lastCheckTime?: string
}

export function WorkspaceFeature({
  initialWorkspaces = [],
  lastCheckTime = '10 min ago',
}: Readonly<WorkspaceFeatureProps>) {
  const router = useRouter()
  const { isLoading, error, deleteWorkspace, createWorkspace, updateWorkspace } = useWorkspaces()

  const [searchQuery, setSearchQuery] = useState('')
  const [workspaces, setWorkspaces] = useState<Workspace[]>(initialWorkspaces)
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
  const [editingWorkspace, setEditingWorkspace] = useState<Workspace | null>(null)
  const [deletingWorkspace, setDeletingWorkspace] = useState<Workspace | null>(null)

  // Use server-side initial workspaces
  const displayWorkspaces = workspaces

  const filteredWorkspaces = useMemo(() => {
    return displayWorkspaces.filter((workspace) => {
      return workspace.name.toLowerCase().includes(searchQuery.toLowerCase())
    })
  }, [
    displayWorkspaces,
    searchQuery,
  ])

  const handleSelectWorkspace = (id: string) => {
    router.push(`/workspaces/${id}`)
  }

  const handleCreateWorkspace = () => {
    setEditingWorkspace(null)
    setIsCreateDialogOpen(true)
  }

  const handleEditWorkspace = (workspace: Workspace) => {
    setEditingWorkspace(workspace)
    setIsCreateDialogOpen(true)
  }

  const handleDialogSubmit = async (data: {
    name: string
    type: 'Personal' | 'Team' | 'Competitor'
  }) => {
    if (editingWorkspace) {
      try {
        const result = await updateWorkspace(editingWorkspace.id, data)
        if (result) {
          setWorkspaces((prev) => prev.map((ws) => (ws.id === editingWorkspace.id ? result : ws)))
          setIsCreateDialogOpen(false)
          router.refresh()
        }
      } catch (err) {
        console.error('Failed to update workspace:', err)
      }
    } else {
      try {
        const result = await createWorkspace(data)
        if (result) {
          setWorkspaces((prev) => [
            result,
            ...prev,
          ])
          setIsCreateDialogOpen(false)
          router.refresh()
        }
      } catch (err) {
        console.error('Failed to create workspace:', err)
      }
    }
  }

  const handleDeleteWorkspace = async (id: string) => {
    const workspace = workspaces.find((ws) => ws.id === id)
    if (workspace) {
      setDeletingWorkspace(workspace)
    }
  }

  const handleConfirmDelete = async () => {
    if (!deletingWorkspace) return
    try {
      await deleteWorkspace(deletingWorkspace.id)
      setWorkspaces((prev) => prev.filter((ws) => ws.id !== deletingWorkspace.id))
      setDeletingWorkspace(null)
    } catch (err) {
      console.error('Failed to delete workspace:', err)
    }
  }

  const handleSettings = () => {
    console.log('Settings')
  }

  return (
    <div className="flex flex-col gap-1">
      {/* Create/Edit Workspace Dialog */}
      <CreateWorkspaceDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
        onSubmit={handleDialogSubmit}
        isLoading={isLoading}
        error={error}
        initialData={editingWorkspace}
        mode={editingWorkspace ? 'edit' : 'create'}
      />

      {/* Delete Workspace Dialog */}
      <DeleteWorkspaceDialog
        open={!!deletingWorkspace}
        onOpenChange={(open) => {
          if (!open) setDeletingWorkspace(null)
        }}
        onConfirm={handleConfirmDelete}
        workspaceName={deletingWorkspace?.name}
        isLoading={isLoading}
      />

      {/* Welcome Container */}
      <WelcomeContainer onSettings={handleSettings} />

      {/* Search and Create */}
      <div className="flex flex-col md:flex-row justify-between items-center self-stretch gap-4 px-4 md:px-8 lg:px-24 py-8">
        <SearchBar value={searchQuery} onChange={setSearchQuery} />
        <div className="flex items-center gap-4 w-full md:w-auto">
          <Button
            variant="default"
            onClick={handleCreateWorkspace}
            className="h-9 px-4 gap-2 bg-primary w-full md:w-auto"
          >
            <SquarePlus className="w-4 h-4" />
            Create workplace
          </Button>
        </div>
      </div>

      {/* Tabs and Last Check */}
      <div className="flex justify-between self-stretch gap-1 px-4 md:px-8 lg:px-24 py-8 pb-1">
        <div />
        <div className="flex justify-center items-center gap-1 p-3">
          <p className="text-sm font-normal text-center text-muted-foreground">
            Last check: {lastCheckTime}
          </p>
        </div>
      </div>

      {/* Workspaces Grid */}
      <div className="flex items-center self-stretch gap-6 px-4 md:px-8 lg:px-24 py-6">
        {isLoading && (
          <div className="flex items-center justify-center w-full">
            <p className="text-muted-foreground">Loading workspaces...</p>
          </div>
        )}

        {error && (
          <div className="flex items-center justify-center w-full">
            <p className="text-destructive">Error loading workspaces: {error.message}</p>
          </div>
        )}

        {!isLoading && !error && filteredWorkspaces.length === 0 ? (
          <EmptyState
            message={searchQuery ? `No workspaces found for "${searchQuery}"` : 'No workspaces'}
          />
        ) : (
          <div className="flex flex-wrap gap-6">
            {filteredWorkspaces.map((workspace) => (
              <WorkspaceCard
                key={workspace.id}
                workspace={{
                  ...workspace,
                  pageCount: 0,
                  status: 'Active',
                }}
                onSelect={handleSelectWorkspace}
                onOpen={() => handleSelectWorkspace(workspace.id)}
                onRename={() => handleEditWorkspace(workspace)}
                onEditTag={() => handleEditWorkspace(workspace)}
                onDelete={() => handleDeleteWorkspace(workspace.id)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

// Re-export types for convenience
export type { Workspace } from './domain/types'
