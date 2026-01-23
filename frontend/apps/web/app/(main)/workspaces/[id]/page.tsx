import { WorkspaceDetailFeature } from '@/features/workspace-detail'
import { WorkspaceBreadcrumbs } from './workspace-breadcrumbs'
import type { BreadcrumbItem } from '@workspace/ui/components/molecules'
import { getWorkspaceServer } from '@/features/workspace/application/services/server'
import { getPagesServer } from '@/features/page/application/services/server'

interface WorkspaceDetailPageProps {
  params: Promise<{
    id: string
  }>
}

export default async function WorkspaceDetailPage({ params }: WorkspaceDetailPageProps) {
  const { id } = await params

  // Fetch workspace and pages from backend
  const workspace = await getWorkspaceServer(id)
  const pages = await getPagesServer(id)

  const breadcrumbs: BreadcrumbItem[] = [
    {
      label: 'Workspaces',
      href: '/workspaces',
    },
    {
      label: workspace.name,
      isCurrent: true,
    },
  ]

  return (
    <>
      <WorkspaceBreadcrumbs breadcrumbs={breadcrumbs} />
      <WorkspaceDetailFeature workspace={workspace} initialPages={pages} />
    </>
  )
}
