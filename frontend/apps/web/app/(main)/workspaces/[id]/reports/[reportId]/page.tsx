import { ReportApi } from '@workspace/services'
import type { BreadcrumbItem } from '@workspace/ui/components/molecules'
import { Badge } from '@workspace/ui/components/atoms/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@workspace/ui/components/atoms/card'
import { ExternalLink, ArrowLeft } from 'lucide-react'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import { getWorkspaceServer } from '@/features/workspace/application/services/server'
import { WorkspaceBreadcrumbs } from '../../workspace-breadcrumbs'

interface ReportDetailPageProps {
  params: Promise<{ id: string; reportId: string }>
}

export default async function ReportDetailPage({ params }: ReportDetailPageProps) {
  const { id: workspaceId, reportId } = await params

  const workspace = await getWorkspaceServer(workspaceId)

  let report
  try {
    report = await ReportApi.getReport(reportId)
  } catch {
    notFound()
  }

  const reportDate = new Date(report.reportDate).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })

  const breadcrumbs: BreadcrumbItem[] = [
    { label: 'Workspaces', href: '/workspaces' },
    { label: workspace.name, href: `/workspaces/${workspaceId}` },
    { label: 'Reports', href: `/workspaces/${workspaceId}/reports` },
    { label: report.title, isCurrent: true },
  ]

  return (
    <>
      <WorkspaceBreadcrumbs breadcrumbs={breadcrumbs} />
      <div className="px-4 md:px-8 lg:px-24 py-8">
        <Link
          href={`/workspaces/${workspaceId}/reports`}
          className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-6"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to reports
        </Link>

        <div className="mb-6">
          <h1 className="text-2xl font-bold text-foreground">{report.title}</h1>
          <p className="text-sm text-muted-foreground mt-1">{reportDate}</p>
        </div>

        {report.pdfUrl && (
          <div className="mb-6">
            <a
              href={report.pdfUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2"
            >
              <Badge variant="secondary" className="gap-1">
                <ExternalLink className="w-3 h-3" />
                Download PDF
              </Badge>
            </a>
          </div>
        )}

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Content</CardTitle>
          </CardHeader>
          <CardContent>
            {Object.keys(report.content).length > 0 ? (
              <pre className="text-sm whitespace-pre-wrap font-mono bg-muted p-4 rounded-md overflow-auto">
                {JSON.stringify(report.content, null, 2)}
              </pre>
            ) : (
              <p className="text-muted-foreground">No content available.</p>
            )}
          </CardContent>
        </Card>
      </div>
    </>
  )
}
