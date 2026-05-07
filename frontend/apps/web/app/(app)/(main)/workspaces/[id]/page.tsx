import { Suspense } from 'react'
import type { BreadcrumbItem } from '@workspace/ui/components/molecules'
import { getPagesServer } from '@/features/page/application/services/server'
import { getWorkspaceServer } from '@/features/workspace/application/services/server'
import { WorkspaceDetailFeature } from '@/features/workspace-detail'
import { WorkspaceDetailSkeleton } from '@/features/workspace-detail/ui/workspace-detail-skeleton'
import { WorkspaceBreadcrumbs } from './workspace-breadcrumbs'

interface WorkspaceDetailPageProps {
  params: Promise<{
    id: string
  }>
}

async function WorkspaceDetailLoader({ id }: { id: string }) {
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

export default async function WorkspaceDetailPage({ params }: WorkspaceDetailPageProps) {
  const { id } = await params

  return (
    <>
      <WorkspaceBreadcrumbs breadcrumbs={[
        { label: 'Workspaces', href: '/workspaces' },
        { label: <span className="inline-block h-4 w-24 bg-muted rounded animate-pulse align-middle" />, isCurrent: true },
      ]} />
      <Suspense fallback={<WorkspaceDetailSkeleton />}>
        <WorkspaceDetailLoader id={id} />
      </Suspense>
    </>
  )
}
