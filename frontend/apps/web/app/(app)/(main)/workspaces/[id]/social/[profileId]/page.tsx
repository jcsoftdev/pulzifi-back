import { Suspense } from 'react'
import { SocialApi } from '@workspace/services/social-api'
import { SocialProfileDetailFeature } from '@/features/social'
import { SocialProfileDetailSkeleton } from '@/features/social/ui/social-profile-detail-skeleton'
import { WorkspaceBreadcrumbs } from '@/app/(app)/(main)/workspaces/[id]/workspace-breadcrumbs'
import { getWorkspaceServer } from '@/features/workspace/application/services/server'

interface SocialProfileDetailPageProps {
  params: Promise<{
    id: string
    profileId: string
  }>
}

async function SocialProfileDetailLoader({
  workspaceId,
  profileId,
}: {
  workspaceId: string
  profileId: string
}) {
  const [workspace, detail, changes, snapshots] = await Promise.all([
    getWorkspaceServer(workspaceId),
    SocialApi.getProfile(profileId),
    SocialApi.listChanges(profileId),
    SocialApi.listSnapshots(profileId),
  ])

  const prevSnapshotSummary = snapshots.find(
    (s) => s.status === 'success' && s.id !== detail.latestSnapshot?.id
  )
  const prevSnapshot = prevSnapshotSummary
    ? await SocialApi.getSnapshot(prevSnapshotSummary.id).catch(() => null)
    : null

  return (
    <>
      <WorkspaceBreadcrumbs
        breadcrumbs={[
          { label: 'Workspaces', href: '/workspaces' },
          { label: workspace.name, href: `/workspaces/${workspaceId}` },
          { label: `@${detail.profile.handle}`, isCurrent: true },
        ]}
      />
      <SocialProfileDetailFeature
        workspaceId={workspaceId}
        profileId={profileId}
        initialProfile={detail.profile}
        initialSnapshot={detail.latestSnapshot}
        initialChanges={changes}
        initialSnapshots={snapshots}
        initialPrevSnapshot={prevSnapshot}
      />
    </>
  )
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
          {
            label: <span className="inline-block h-3.5 w-24 bg-muted rounded animate-pulse" />,
            href: `/workspaces/${workspaceId}`,
          },
          {
            label: <span className="inline-block h-3.5 w-28 bg-muted rounded animate-pulse" />,
            isCurrent: true,
          },
        ]}
      />
      <Suspense fallback={<SocialProfileDetailSkeleton />}>
        <SocialProfileDetailLoader workspaceId={workspaceId} profileId={profileId} />
      </Suspense>
    </>
  )
}
