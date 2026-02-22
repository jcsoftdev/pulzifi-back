import type { BreadcrumbItem } from '@workspace/ui/components/molecules'
import { getPagesServer } from '@/features/page/application/services/server'
import { getWorkspaceServer } from '@/features/workspace/application/services/server'
import { ReportsFeature } from '@/features/reports'
import { WorkspaceBreadcrumbs } from '../workspace-breadcrumbs'

interface WorkspaceReportsPageProps {
  params: Promise<{ id: string }>
}

export default async function WorkspaceReportsPage({ params }: WorkspaceReportsPageProps) {
  const { id } = await params

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
