'use client'

import { type Page, PageApi } from '@workspace/services/page-api'
import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { notification } from '@/lib/notification'
import { DeletePageDialog } from '@/features/page/ui/delete-page-dialog'
import { EditPageDialog } from '@/features/page/ui/edit-page-dialog'
import type { EditPageDto } from '@/features/page/domain/types'
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
  const [isRunning, setIsRunning] = useState(false)
  const [editError, setEditError] = useState<Error | null>(null)

  const handleUpdatePage = async (id: string, data: EditPageDto) => {
    setIsActionLoading(true)
    setEditError(null)
    try {
      // 1. Update core page fields (name, url, tags)
      const updated = await PageApi.updatePage(id, {
        name: data.name,
        url: data.url,
        tags: data.tags,
      })

      // 2. Update monitoring config (frequency, schedule, insights, alerts, selector type)
      await PageApi.updateMonitoringConfig(id, {
        checkFrequency: data.checkFrequency,
        blockAdsCookies: data.blockAdsCookies,
        scheduleType: data.scheduleType,
        enabledInsightTypes: data.enabledInsightTypes,
        enabledAlertConditions: data.enabledAlertConditions,
        customAlertCondition: data.customAlertCondition,
        selectorType: data.selectorType,
        cssSelector: data.cssSelector,
        xpathSelector: data.xpathSelector,
        selectorOffsets: data.selectorOffsets,
      })

      // 3. Save sections if they were provided (modified via selector UI)
      if (data.sections !== undefined) {
        if (data.sections.length > 0) {
          await PageApi.saveSections(
            id,
            data.sections.map((s, i) => ({
              name: s.name,
              cssSelector: s.cssSelector,
              xpathSelector: s.xpathSelector,
              selectorOffsets: s.selectorOffsets,
              sortOrder: s.sortOrder ?? i,
            }))
          )
        } else {
          // Sections array is empty — user cleared sections (switched to full_page)
          // The saveSections with empty array will replace all existing sections
          await PageApi.saveSections(id, [])
        }
      }

      setPage(updated)
      setIsEditOpen(false)
      router.refresh()
      notification.action({
        title: 'Page updated',
        description: `"${updated.name}" has been updated.`,
      })
    } catch (err) {
      console.error('Failed to update page:', err)
      const error = err instanceof Error ? err : new Error('Please try again.')
      setEditError(error)
      notification.error({
        title: 'Failed to update page',
        description: error.message,
      })
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
      notification.error({
        title: 'Failed to delete page',
        description: err instanceof Error ? err.message : 'Please try again.',
      })
      setIsActionLoading(false)
    }
  }

  // Listen for SSE-driven check status changes from ChecksHistory.
  useEffect(() => {
    const handler = (e: Event) => {
      const hasActive = (e as CustomEvent<boolean>).detail
      if (!hasActive) setIsRunning(false)
    }
    window.addEventListener('checks:active', handler)
    return () => window.removeEventListener('checks:active', handler)
  }, [])

  const handleRunNow = async () => {
    setIsRunning(true)
    try {
      await PageApi.triggerCheck(page.id)
      notification.success({ title: 'Check triggered', description: 'A new check is running now.' })
      // Signal ChecksHistory to refresh (fallback in case SSE is delayed).
      window.dispatchEvent(new Event('checks:refresh'))
    } catch (err) {
      console.error('Failed to trigger check:', err)
      notification.error({
        title: 'Failed to run check',
        description: err instanceof Error ? err.message : 'Please try again.',
      })
      setIsRunning(false)
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
        onRunNow={handleRunNow}
        isRunning={isRunning}
      />

      <EditPageDialog
        open={isEditOpen}
        isLoading={isActionLoading}
        page={page}
        onOpenChange={setIsEditOpen}
        onSubmit={handleUpdatePage}
        error={editError}
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
