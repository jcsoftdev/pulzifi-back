'use client'

import type { Page } from '@workspace/services'
import { Plus } from 'lucide-react'
import { useCallback, useState } from 'react'
import { notification } from '@/lib/notification'
import { Button } from '@workspace/ui/components/atoms/button'
import { useReports } from './application/use-reports'
import { CreateReportDialog } from './ui/create-report-dialog'
import { ReportsTable } from './ui/reports-table'

interface ReportsFeatureProps {
  workspaceId: string
  pages: Page[]
}

export function ReportsFeature({ workspaceId, pages }: Readonly<ReportsFeatureProps>) {
  const pageIds = pages.map((p) => p.id)
  const { reports, loading, createReport } = useReports(pageIds)

  const [createOpen, setCreateOpen] = useState(false)
  const [actionError, setActionError] = useState<Error | null>(null)
  const [actionLoading, setActionLoading] = useState(false)

  const handleCreate = useCallback(
    async (data: { pageId: string; title: string; reportDate: string }) => {
      setActionError(null)
      setActionLoading(true)
      try {
        await createReport({
          pageId: data.pageId,
          title: data.title,
          reportDate: data.reportDate,
        })
        setCreateOpen(false)
        notification.success({ title: 'Report created', description: `"${data.title}" has been created.` })
      } catch (err) {
        setActionError(err instanceof Error ? err : new Error('Failed to create report'))
        notification.error({ title: 'Failed to create report', description: err instanceof Error ? err.message : 'Please try again.' })
      } finally {
        setActionLoading(false)
      }
    },
    [createReport]
  )

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
      <ReportsTable reports={reports} loading={loading} workspaceId={workspaceId} />

      {/* Dialogs */}
      <CreateReportDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        onSubmit={handleCreate}
        pages={pages}
        isLoading={actionLoading}
        error={actionError}
      />
    </div>
  )
}
