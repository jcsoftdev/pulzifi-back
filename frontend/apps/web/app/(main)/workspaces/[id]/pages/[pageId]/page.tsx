import { Suspense } from 'react'
import { WorkspaceBreadcrumbs } from '@/app/(main)/workspaces/[id]/workspace-breadcrumbs'
import { PageDetailFeature } from '@/features/page-detail'
import { PageDetailSkeleton } from '@/features/page-detail/ui/page-detail-skeleton'

interface PageProps {
  params: Promise<{
    id: string
    pageId: string
  }>
}

export default async function PageDetailPage(props: Readonly<PageProps>) {
  const params = await props.params
  const { id: workspaceId, pageId } = params

  return (
    <>
      <WorkspaceBreadcrumbs breadcrumbs={[
        { label: 'Workspaces', href: '/workspaces' },
        { label: <span className="inline-block h-4 w-24 bg-muted rounded animate-pulse align-middle" />, href: `/workspaces/${workspaceId}` },
        { label: <span className="inline-block h-4 w-32 bg-muted rounded animate-pulse align-middle" />, isCurrent: true },
      ]} />
      <Suspense fallback={<PageDetailSkeleton />}>
        <PageDetailFeature workspaceId={workspaceId} pageId={pageId} />
      </Suspense>
    </>
  )
}
