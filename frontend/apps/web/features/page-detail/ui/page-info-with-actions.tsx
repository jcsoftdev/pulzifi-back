'use client'

import { type Page, PageApi } from '@workspace/services/page-api'
import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { notification } from '@/lib/notification'
import { DeletePageDialog } from '@/features/page/ui/delete-page-dialog'
import { EditPageDialog } from '@/features/page/ui/edit-page-dialog'
import { PageInfoCard } from './page-info-card'

interface PageInfoWithActionsProps {
  initialPage: Page
  workspaceId: string
}

export function PageInfoWithActions({
  initialPage,
  workspaceId,
}: Readonly<PageInfoWithActionsProps>) {
  const router = useRouter()
  const [page, setPage] = useState<Page>(initialPage)
  const [isEditOpen, setIsEditOpen] = useState(false)
  const [isDeleteOpen, setIsDeleteOpen] = useState(false)
  const [isActionLoading, setIsActionLoading] = useState(false)

  const handleUpdatePage = async (
    id: string,
    data: {
      name: string
      url: string
    }
  ) => {
    setIsActionLoading(true)
    try {
      const updated = await PageApi.updatePage(id, data)
      setPage(updated)
      setIsEditOpen(false)
      router.refresh()
      notification.action({ title: 'Page updated', description: `"${updated.name}" has been updated.` })
    } catch (err) {
      console.error('Failed to update page:', err)
      notification.error({ title: 'Failed to update page', description: err instanceof Error ? err.message : 'Please try again.' })
    } finally {
      setIsActionLoading(false)
    }
  }

  const handleDeletePage = async () => {
    setIsActionLoading(true)
    try {
      await PageApi.deletePage(page.id)
      notification.success({ title: 'Page deleted' })
      router.push(`/workspaces/${workspaceId}`)
    } catch (err) {
      console.error('Failed to delete page:', err)
      notification.error({ title: 'Failed to delete page', description: err instanceof Error ? err.message : 'Please try again.' })
      setIsActionLoading(false)
    }
  }

  const handleViewChanges = () => {
    router.push(`/workspaces/${workspaceId}/pages/${page.id}/changes`)
  }

  return (
    <>
      <PageInfoCard
        page={page}
        onEdit={() => setIsEditOpen(true)}
        onDelete={() => setIsDeleteOpen(true)}
        onViewChanges={handleViewChanges}
      />

      <EditPageDialog
        open={isEditOpen}
        isLoading={isActionLoading}
        page={page}
        onOpenChange={setIsEditOpen}
        onSubmit={handleUpdatePage}
      />

      <DeletePageDialog
        open={isDeleteOpen}
        isLoading={isActionLoading}
        pageName={page.name}
        onOpenChange={setIsDeleteOpen}
        onConfirm={handleDeletePage}
      />
    </>
  )
}
