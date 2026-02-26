import { Suspense } from 'react'
import type { BreadcrumbItem } from '@workspace/ui/components/molecules'
import { getPagesServer } from '@/features/page/application/services/server'
import { getWorkspaceServer } from '@/features/workspace/application/services/server'
import { ReportsFeature } from '@/features/reports'
import { WorkspaceBreadcrumbs } from '../workspace-breadcrumbs'

interface WorkspaceReportsPageProps {
  params: Promise<{ id: string }>
}

async function ReportsLoader({ id }: { id: string }) {
  const workspace = await getWorkspaceServer(id)
  const pages = await getPagesServer(id)

  const breadcrumbs: BreadcrumbItem[] = [
    { label: 'Workspaces', href: '/workspaces' },
    { label: workspace.name, href: `/workspaces/${id}` },
    { label: 'Reports', isCurrent: true },
  ]

  return (
    <>
      <WorkspaceBreadcrumbs breadcrumbs={breadcrumbs} />
      <ReportsFeature workspaceId={id} pages={pages} />
    </>
  )
}

function ReportsSkeleton() {
  return (
    <div className="px-4 md:px-8 lg:px-24 py-8 space-y-4">
      <div className="h-64 bg-muted rounded-lg animate-pulse" />
    </div>
  )
}

export default async function WorkspaceReportsPage({ params }: WorkspaceReportsPageProps) {
  const { id } = await params

  return (
    <>
      <WorkspaceBreadcrumbs breadcrumbs={[
        { label: 'Workspaces', href: '/workspaces' },
        { label: <span className="inline-block h-4 w-24 bg-muted rounded animate-pulse align-middle" />, href: `/workspaces/${id}` },
        { label: 'Reports', isCurrent: true },
      ]} />
      <Suspense fallback={<ReportsSkeleton />}>
        <ReportsLoader id={id} />
      </Suspense>
    </>
  )
}
