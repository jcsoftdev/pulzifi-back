'use client'
import { notification } from '@/lib/notification'
import { PageApi } from '@workspace/services/page-api'
import { Badge } from '@workspace/ui/components/atoms/badge'
import { Button } from '@workspace/ui/components/atoms/button'
import { Input } from '@workspace/ui/components/atoms/input'
import { FileText, Settings, SquarePlus, Trash2 } from 'lucide-react'
import { useRouter } from 'next/navigation'

import { useState } from 'react'
import type { CreatePageDto, Page } from '@/features/page/domain/types'
import { AddPageDialog } from '@/features/page/ui/add-page-dialog'
import { BulkDeletePagesDialog } from '@/features/page/ui/bulk-delete-pages-dialog'
import { BulkFrequencyChangeDialog } from '@/features/page/ui/bulk-frequency-change-dialog'
import { DeletePageDialog } from '@/features/page/ui/delete-page-dialog'
import { EditPageDialog } from '@/features/page/ui/edit-page-dialog'
import { PagesTable } from '@/features/page/ui/pages-table'
import { useWorkspaces } from '@/features/workspace/application/hooks/use-workspaces'
import type { Workspace, WorkspaceType } from '@/features/workspace/domain/types'
import { DeleteWorkspaceDialog } from '@/features/workspace/ui/delete-workspace-dialog'
import { EditWorkspaceDialog } from '@/features/workspace/ui/edit-workspace-dialog'

export interface WorkspaceDetailContentProps {
  workspace: Workspace
  initialPages?: Page[]
}

export function WorkspaceDetailContent({
  workspace: initialWorkspace,
  initialPages = [],
}: Readonly<WorkspaceDetailContentProps>) {
  const router = useRouter()
  const { updateWorkspace, deleteWorkspace, isLoading: isWorkspaceLoading } = useWorkspaces()

  const [workspace, setWorkspace] = useState<Workspace>(initialWorkspace)
  const [pages, setPages] = useState<Page[]>(initialPages)
  const [searchQuery, setSearchQuery] = useState('')
  const [isAddPageOpen, setIsAddPageOpen] = useState(false)
  const [isEditWorkspaceOpen, setIsEditWorkspaceOpen] = useState(false)
  const [isDeleteWorkspaceOpen, setIsDeleteWorkspaceOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const [isEditPageOpen, setIsEditPageOpen] = useState(false)
  const [isDeletePageOpen, setIsDeletePageOpen] = useState(false)
  const [selectedPage, setSelectedPage] = useState<Page | null>(null)

  const [isBulkDeleteOpen, setIsBulkDeleteOpen] = useState(false)
  const [pendingBulkDeleteIds, setPendingBulkDeleteIds] = useState<string[]>([])

  const [isBulkFrequencyOpen, setIsBulkFrequencyOpen] = useState(false)
  const [pendingBulkFrequency, setPendingBulkFrequency] = useState<{ ids: string[]; frequency: string } | null>(null)

  const handleAddPage = async (data: CreatePageDto) => {
    setIsLoading(true)
    setError(null)

    try {
      const newPage = await PageApi.createPage(data)
      // Seed default monitoring config: Weekdays schedule, frequency Off
      await PageApi.updateMonitoringConfig(newPage.id, {
        scheduleType: 'all_time',
        checkFrequency: 'Off',
      })
      setPages((prev) => [
        newPage,
        ...prev,
      ])
      setIsAddPageOpen(false)
      notification.success({ title: 'Page added', description: `"${newPage.name}" has been added.` })
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to add page'))
      notification.error({ title: 'Failed to add page', description: err instanceof Error ? err.message : 'Please try again.' })
    } finally {
      setIsLoading(false)
    }
  }

  const handleEditPageClick = (page: Page) => {
    setSelectedPage(page)
    setIsEditPageOpen(true)
  }

  const handleDeletePageClick = (page: Page) => {
    setSelectedPage(page)
    setIsDeletePageOpen(true)
  }

  const handleUpdatePage = async (
    pageId: string,
    data: {
      name: string
      url: string
    }
  ) => {
    setIsLoading(true)
    try {
      const updatedPage = await PageApi.updatePage(pageId, data)
      setPages((prev) => prev.map((p) => (p.id === pageId ? updatedPage : p)))
      setIsEditPageOpen(false)
      setSelectedPage(null)
      notification.success({ title: 'Page updated', description: `"${updatedPage.name}" has been updated.` })
    } catch (err) {
      console.error('Failed to update page:', err)
      notification.error({ title: 'Failed to update page', description: err instanceof Error ? err.message : 'Please try again.' })
    } finally {
      setIsLoading(false)
    }
  }

  const handleDeletePage = async () => {
    if (!selectedPage) return
    setIsLoading(true)
    try {
      await PageApi.deletePage(selectedPage.id)
      setPages((prev) => prev.filter((p) => p.id !== selectedPage.id))
      setIsDeletePageOpen(false)
      setSelectedPage(null)
      notification.success({ title: 'Page deleted' })
    } catch (err) {
      console.error('Failed to delete page:', err)
      notification.error({ title: 'Failed to delete page', description: err instanceof Error ? err.message : 'Please try again.' })
    } finally {
      setIsLoading(false)
    }
  }

  const handleUpdateWorkspace = async (
    id: string,
    data: {
      name: string
      type: WorkspaceType
      tags: string[]
    }
  ) => {
    try {
      const updated = await updateWorkspace(id, data)
      if (updated) {
        setWorkspace(updated)
        setIsEditWorkspaceOpen(false)
        router.refresh()
        notification.success({ title: 'Workspace updated', description: `"${updated.name}" has been updated.` })
      }
    } catch (err) {
      console.error('Failed to update workspace:', err)
      notification.error({ title: 'Failed to update workspace', description: err instanceof Error ? err.message : 'Please try again.' })
    }
  }

  const handleDeleteWorkspace = async () => {
    try {
      await deleteWorkspace(workspace.id)
      notification.success({ title: 'Workspace deleted' })
      router.push('/workspaces')
    } catch (err) {
      console.error('Failed to delete workspace:', err)
      notification.error({ title: 'Failed to delete workspace', description: err instanceof Error ? err.message : 'Please try again.' })
    }
  }

  const handleViewChanges = (pageId: string) => {
    router.push(`/workspaces/${workspace.id}/pages/${pageId}/changes`)
  }

  const handlePageClick = (pageId: string) => {
    router.push(`/workspaces/${workspace.id}/pages/${pageId}`)
  }

  const handleCheckFrequencyChange = async (pageId: string, frequency: string) => {
    setPages((prev) =>
      prev.map((page) =>
        page.id === pageId
          ? {
              ...page,
              checkFrequency: frequency,
            }
          : page
      )
    )

    try {
      await PageApi.updateMonitoringConfig(pageId, {
        checkFrequency: frequency,
      })
      notification.success({ title: 'Check frequency updated' })
    } catch (err) {
      setPages(initialPages)
      console.error('Failed to update check frequency:', err)
      notification.error({ title: 'Failed to update check frequency', description: err instanceof Error ? err.message : 'Please try again.' })
    }
  }

  const handleBulkDelete = (pageIds: string[]) => {
    setPendingBulkDeleteIds(pageIds)
    setIsBulkDeleteOpen(true)
  }

  const handleBulkDeleteConfirm = async () => {
    try {
      await PageApi.bulkDeletePages(pendingBulkDeleteIds)
      setPages((prev) => prev.filter((p) => !pendingBulkDeleteIds.includes(p.id)))
      notification.success({ title: `${pendingBulkDeleteIds.length} page${pendingBulkDeleteIds.length > 1 ? 's' : ''} deleted` })
    } catch (err) {
      console.error('Failed to bulk delete pages:', err)
      notification.error({ title: 'Failed to delete pages', description: err instanceof Error ? err.message : 'Please try again.' })
    } finally {
      setPendingBulkDeleteIds([])
    }
  }

  const handleBulkFrequencyChange = (pageIds: string[], frequency: string) => {
    setPendingBulkFrequency({ ids: pageIds, frequency })
    setIsBulkFrequencyOpen(true)
  }

  const handleBulkFrequencyConfirm = async () => {
    if (!pendingBulkFrequency) return
    const { ids, frequency } = pendingBulkFrequency
    setPages((prev) =>
      prev.map((page) => (ids.includes(page.id) ? { ...page, checkFrequency: frequency } : page))
    )
    try {
      await PageApi.bulkUpdateFrequency(ids, frequency)
      notification.success({ title: 'Check frequency updated for selected pages' })
    } catch (err) {
      setPages(initialPages)
      console.error('Failed to bulk update check frequency:', err)
      notification.error({ title: 'Failed to update check frequency', description: err instanceof Error ? err.message : 'Please try again.' })
    } finally {
      setPendingBulkFrequency(null)
    }
  }

  const filteredPages = pages.filter((page) =>
    page.name.toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <div className="flex-1 flex flex-col bg-background">
      <div className="flex flex-col md:flex-row justify-between items-start gap-4 px-4 md:px-8 lg:px-24 py-6">
        <div className="flex flex-col gap-2">
          <div className="flex flex-wrap items-center gap-3">
            <h1 className="text-2xl font-semibold text-foreground">
              Added pages for {workspace.name}
            </h1>
            <div className="flex gap-1">
              {workspace.tags?.map((tag) => (
                <Badge key={tag} variant="secondary">
                  {tag}
                </Badge>
              ))}
            </div>
          </div>
          <p className="text-base font-normal text-muted-foreground">
            Here are all the pages you've added to this workspace.
          </p>
        </div>

        <div className="flex items-center gap-2 w-full md:w-auto">
          <Button
            variant="outline"
            onClick={() => router.push(`/workspaces/${workspace.id}/reports`)}
            className="gap-2 flex-1 md:flex-none"
          >
            <FileText className="w-4 h-4" />
            Reports
          </Button>
          <Button
            variant="outline"
            onClick={() => setIsEditWorkspaceOpen(true)}
            className="gap-2 flex-1 md:flex-none"
          >
            <Settings className="w-4 h-4" />
            Edit Workspace
          </Button>
          <Button
            variant="destructive"
            onClick={() => setIsDeleteWorkspaceOpen(true)}
            size="icon"
            className="h-9 w-9 shrink-0"
          >
            <Trash2 className="w-4 h-4" />
          </Button>
        </div>
      </div>

      <div className="flex flex-col md:flex-row justify-between items-stretch md:items-center px-4 md:px-8 lg:px-24 py-2 gap-4">
        <div className="relative flex-1 w-full md:max-w-sm">
          <svg
            width="17"
            height="17"
            viewBox="0 0 17 17"
            fill="none"
            className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
          >
            <title>Search</title>
            <path
              d="M7.79167 13.4583C10.8292 13.4583 13.2917 10.9958 13.2917 7.95833C13.2917 4.92084 10.8292 2.45833 7.79167 2.45833C4.75418 2.45833 2.29167 4.92084 2.29167 7.95833C2.29167 10.9958 4.75418 13.4583 7.79167 13.4583Z"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <path
              d="M14.5833 14.75L11.7292 11.8958"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          <Input
            type="search"
            placeholder="Search pages"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>

        <div className="flex items-center gap-4">
          <Button
            variant="default"
            onClick={() => setIsAddPageOpen(true)}
            className="h-9 px-4 gap-2 bg-primary w-full md:w-auto"
          >
            <SquarePlus className="w-4 h-4" />
            Add page
          </Button>
        </div>
      </div>

      <div className="px-4 md:px-8 lg:px-24 py-2 pb-6">
        <PagesTable
          pages={filteredPages}
          onViewChanges={handleViewChanges}
          onPageClick={handlePageClick}
          onCheckFrequencyChange={handleCheckFrequencyChange}
          onEdit={handleEditPageClick}
          onDelete={handleDeletePageClick}
          onBulkDelete={handleBulkDelete}
          onBulkFrequencyChange={handleBulkFrequencyChange}
        />
      </div>

      <AddPageDialog
        open={isAddPageOpen}
        onOpenChange={setIsAddPageOpen}
        onSubmit={handleAddPage}
        workspaceId={workspace.id}
        isLoading={isLoading}
        error={error}
      />

      <EditPageDialog
        open={isEditPageOpen}
        onOpenChange={setIsEditPageOpen}
        onSubmit={handleUpdatePage}
        page={selectedPage}
        isLoading={isLoading}
      />

      <DeletePageDialog
        open={isDeletePageOpen}
        onOpenChange={setIsDeletePageOpen}
        onConfirm={handleDeletePage}
        pageName={selectedPage?.name ?? ''}
        isLoading={isLoading}
      />

      <BulkDeletePagesDialog
        open={isBulkDeleteOpen}
        onOpenChange={setIsBulkDeleteOpen}
        onConfirm={handleBulkDeleteConfirm}
        count={pendingBulkDeleteIds.length}
        isLoading={isLoading}
      />

      <BulkFrequencyChangeDialog
        open={isBulkFrequencyOpen}
        onOpenChange={setIsBulkFrequencyOpen}
        onConfirm={handleBulkFrequencyConfirm}
        count={pendingBulkFrequency?.ids.length ?? 0}
        frequency={pendingBulkFrequency?.frequency ?? ''}
        isLoading={isLoading}
      />

      <EditWorkspaceDialog
        open={isEditWorkspaceOpen}
        onOpenChange={setIsEditWorkspaceOpen}
        onSubmit={handleUpdateWorkspace}
        isLoading={isWorkspaceLoading}
        error={null}
        workspace={workspace}
      />

      <DeleteWorkspaceDialog
        open={isDeleteWorkspaceOpen}
        onOpenChange={setIsDeleteWorkspaceOpen}
        onConfirm={handleDeleteWorkspace}
        workspaceName={workspace.name}
        isLoading={isWorkspaceLoading}
      />
    </div>
  )
}
