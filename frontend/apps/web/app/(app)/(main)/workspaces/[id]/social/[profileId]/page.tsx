import { Suspense } from 'react'
import { SocialProfileDetailFeature } from '@/features/social'
import { WorkspaceBreadcrumbs } from '@/app/(app)/(main)/workspaces/[id]/workspace-breadcrumbs'

interface SocialProfileDetailPageProps {
  params: Promise<{
    id: string
    profileId: string
  }>
}

export default async function SocialProfileDetailPage({
  params,
}: SocialProfileDetailPageProps) {
  const { id: workspaceId, profileId } = await params

  return (
    <>
      <WorkspaceBreadcrumbs
        breadcrumbs={[
          { label: 'Workspaces', href: '/workspaces' },
          { label: 'Workspace', href: `/workspaces/${workspaceId}` },
          { label: 'Social profile', isCurrent: true },
        ]}
      />
      <Suspense
        fallback={
          <div className="flex-1 p-8 flex flex-col gap-6 animate-pulse">
            <div className="h-24 bg-muted rounded-lg" />
            <div className="h-48 bg-muted rounded-lg" />
            <div className="h-64 bg-muted rounded-lg" />
          </div>
        }
      >
        <SocialProfileDetailFeature workspaceId={workspaceId} profileId={profileId} />
      </Suspense>
    </>
  )
}
