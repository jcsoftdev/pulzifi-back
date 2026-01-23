'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { PageApi, type Page } from '@workspace/services/page-api'
import { EditPageDialog } from '@/features/page/ui/edit-page-dialog'
import { DeletePageDialog } from '@/features/page/ui/delete-page-dialog'
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
      router.refresh() // Refresh server components to update other parts if needed
    } catch (err) {
      console.error('Failed to update page:', err)
    } finally {
      setIsActionLoading(false)
    }
  }

  const handleDeletePage = async () => {
    setIsActionLoading(true)
    try {
      await PageApi.deletePage(page.id)
      router.push(`/workspaces/${workspaceId}`)
    } catch (err) {
      console.error('Failed to delete page:', err)
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
