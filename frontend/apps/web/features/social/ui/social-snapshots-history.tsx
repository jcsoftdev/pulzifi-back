'use client'

import { Badge } from '@workspace/ui/components/atoms/badge'
import { Skeleton } from '@workspace/ui/components/atoms/skeleton'
import { formatRelativeTime } from '@workspace/ui/lib/date'
import type { SocialSnapshotSummary } from '@workspace/services/social-api'

interface SocialSnapshotsHistoryProps {
  snapshots: SocialSnapshotSummary[]
  isLoading: boolean
  changesBySnapshotId: Map<string, string> // snapshotId -> changeId
  onChangeClick: (changeId: string) => void
}

export function SocialSnapshotsHistory({
  snapshots,
  isLoading,
  changesBySnapshotId,
  onChangeClick,
}: Readonly<SocialSnapshotsHistoryProps>) {
  if (isLoading) {
    return (
      <div className="flex flex-col gap-3">
        {[1, 2, 3].map((i) => (
          <Skeleton key={i} className="h-14 w-full rounded-lg" />
        ))}
      </div>
    )
  }

  if (snapshots.length === 0) {
    return <p className="text-sm text-muted-foreground">No runs recorded yet.</p>
  }

  return (
    <div className="relative border-l border-border ml-2 space-y-6">
      {snapshots.map((snap) => {
        const isSuccess = snap.status === 'success'
        const changeId = changesBySnapshotId.get(snap.id)
        const hasChange = !!changeId

        const dotColor = !isSuccess
          ? 'bg-destructive'
          : hasChange
            ? 'bg-orange-500'
            : 'bg-green-500'

        const row = (
          <div className="flex flex-col gap-0.5">
            <span className="text-sm text-muted-foreground">
              {formatRelativeTime(snap.capturedAt)}
            </span>
            {isSuccess ? (
              <div className="flex items-center gap-2">
                <span className="text-sm text-foreground">
                  {snap.followersCount.toLocaleString()} followers &middot; {snap.postsCount} posts
                </span>
                {hasChange && (
                  <Badge variant="outline" className="text-xs text-orange-500 border-orange-500/50">
                    Change detected
                  </Badge>
                )}
              </div>
            ) : (
              <span className="text-sm text-destructive">
                Failed{snap.error ? `: ${snap.error}` : ''}
              </span>
            )}
          </div>
        )

        return (
          <div key={snap.id} className="relative pl-6">
            <div
              className={`absolute -left-1.5 top-1.5 h-3 w-3 rounded-full border-2 border-background ${dotColor}`}
            />
            {hasChange ? (
              <button
                type="button"
                className="text-left w-full hover:bg-muted/50 -ml-2 pl-2 py-1 rounded-md transition-colors cursor-pointer"
                onClick={() => onChangeClick(changeId)}
              >
                {row}
              </button>
            ) : (
              row
            )}
          </div>
        )
      })}
    </div>
  )
}
