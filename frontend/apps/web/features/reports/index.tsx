'use client'

import type { Page, Report } from '@workspace/services'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@workspace/ui/components/atoms/alert-dialog'
import { Button } from '@workspace/ui/components/atoms/button'
import { Plus } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { useCallback, useState } from 'react'
import { notification } from '@/lib/notification'
import { useReports } from './application/use-reports'
import { CreateReportDialog } from './ui/create-report-dialog'
import { ReportsTable } from './ui/reports-table'

interface ReportsFeatureProps {
  workspaceId: string
  pages: Page[]
}

export function ReportsFeature({ workspaceId, pages }: Readonly<ReportsFeatureProps>) {
  const router = useRouter()
  const pageIds = pages.map((p) => p.id)
  const { reports, loading, createReport, deleteReport } = useReports(pageIds)

  const [createOpen, setCreateOpen] = useState(false)
  const [actionError, setActionError] = useState<Error | null>(null)
  const [actionLoading, setActionLoading] = useState(false)
  const [reportToDelete, setReportToDelete] = useState<Report | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)

  const handleCreate = useCallback(
    async (data: { pageId: string; title: string; reportDate: string }) => {
      setActionError(null)
      setActionLoading(true)
      try {
        const report = await createReport({
          pageId: data.pageId,
          title: data.title,
          reportDate: data.reportDate,
        })
        setCreateOpen(false)
        notification.success({
          title: 'Report generated',
          description: `"${data.title}" is ready.`,
        })
        // Open the freshly generated report instead of leaving the user on the list.
        router.push(`/workspaces/${workspaceId}/reports/${report.id}`)
      } catch (err) {
        setActionError(err instanceof Error ? err : new Error('Failed to create report'))
        notification.error({
          title: 'Failed to generate report',
          description: err instanceof Error ? err.message : 'Please try again.',
        })
      } finally {
        setActionLoading(false)
      }
    },
    [
      createReport,
      router,
      workspaceId,
    ]
  )

  const handleConfirmDelete = useCallback(async () => {
    if (!reportToDelete) return
    setIsDeleting(true)
    try {
      await deleteReport(reportToDelete.id)
      notification.success({
        title: 'Report deleted',
        description: `"${reportToDelete.title}" was removed.`,
      })
      setReportToDelete(null)
    } catch (err) {
      notification.error({
        title: 'Failed to delete report',
        description: err instanceof Error ? err.message : 'Please try again.',
      })
    } finally {
      setIsDeleting(false)
    }
  }, [
    deleteReport,
    reportToDelete,
  ])

  return (
    <div className="px-4 md:px-8 lg:px-24 py-8">
      {/* Page header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Reports</h1>
          <p className="text-sm text-muted-foreground mt-1">View and create monitoring reports</p>
        </div>
        <Button
          onClick={() => {
            setActionError(null)
            setCreateOpen(true)
          }}
          size="sm"
        >
          <Plus className="w-4 h-4 mr-2" />
          New report
        </Button>
      </div>

      {/* Reports list */}
      <ReportsTable
        reports={reports}
        loading={loading}
        workspaceId={workspaceId}
        onDelete={setReportToDelete}
      />

      {/* Dialogs */}
      <CreateReportDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        onSubmit={handleCreate}
        pages={pages}
        isLoading={actionLoading}
        error={actionError}
      />

      <AlertDialog
        open={reportToDelete !== null}
        onOpenChange={(open) => !open && setReportToDelete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete report</AlertDialogTitle>
            <AlertDialogDescription>
              {reportToDelete
                ? `Are you sure you want to delete "${reportToDelete.title}"? This action cannot be undone.`
                : ''}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={(e) => {
                e.preventDefault()
                handleConfirmDelete()
              }}
              disabled={isDeleting}
              className="bg-destructive text-white hover:bg-destructive/90"
            >
              {isDeleting ? 'Deleting…' : 'Delete report'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
