'use client'

import type { Report } from '@workspace/services'
import { Badge } from '@workspace/ui/components/atoms/badge'
import { Button } from '@workspace/ui/components/atoms/button'
import { Card, CardContent } from '@workspace/ui/components/atoms/card'
import { ExternalLink, FileText, Trash2 } from 'lucide-react'
import Link from 'next/link'
import { formatReportDate } from '../domain/types'

interface ReportsTableProps {
  reports: Report[]
  loading: boolean
  workspaceId: string
  onDelete?: (report: Report) => void
}

export function ReportsTable({
  reports,
  loading,
  workspaceId,
  onDelete,
}: Readonly<ReportsTableProps>) {
  if (loading) {
    return (
      <div className="grid gap-4">
        {[
          1,
          2,
          3,
        ].map((n) => (
          <Card key={n}>
            <CardContent className="p-4">
              <div className="flex items-center gap-4">
                <div className="h-5 w-48 bg-muted animate-pulse rounded" />
                <div className="h-4 w-24 bg-muted animate-pulse rounded" />
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    )
  }

  if (reports.length === 0) {
    return (
      <Card>
        <CardContent className="p-8 text-center">
          <FileText className="w-12 h-12 text-muted-foreground mx-auto mb-3" />
          <p className="text-muted-foreground">No reports yet</p>
          <p className="text-sm text-muted-foreground mt-1">
            Create your first report to get started.
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="grid gap-3">
      {reports.map((report) => (
        <Link key={report.id} href={`/workspaces/${workspaceId}/reports/${report.id}`}>
          <Card className="hover:bg-accent/50 transition-colors cursor-pointer">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3 min-w-0">
                  <FileText className="w-5 h-5 text-muted-foreground shrink-0" />
                  <div className="min-w-0">
                    <p className="font-medium truncate">{report.title}</p>
                    <p className="text-sm text-muted-foreground">
                      {formatReportDate(report.reportDate)}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  {report.pdfUrl && (
                    <Badge variant="secondary" className="gap-1">
                      <ExternalLink className="w-3 h-3" />
                      PDF
                    </Badge>
                  )}
                  {onDelete && (
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 text-muted-foreground hover:text-destructive"
                      aria-label={`Delete report ${report.title}`}
                      // Row is a <Link>; don't navigate when deleting.
                      onClick={(e) => {
                        e.preventDefault()
                        e.stopPropagation()
                        onDelete(report)
                      }}
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>
        </Link>
      ))}
    </div>
  )
}
