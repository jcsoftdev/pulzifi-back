'use client'

import { Badge } from '@workspace/ui/components/atoms/badge'
import { Skeleton } from '@workspace/ui/components/atoms/skeleton'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@workspace/ui/components/atoms/sheet'
import { SocialApi } from '@workspace/services/social-api'
import { useEffect, useState } from 'react'
import type { SocialChangeDetail } from '../../domain/types'
import { formatRelativeTime } from '@workspace/ui/lib/date'

interface SocialChangeDetailSheetProps {
  changeId: string | null
  onClose: () => void
}

const CHANGE_TYPE_LABELS: Record<string, string> = {
  new_post: 'New post',
  removed_post: 'Post removed',
  caption_edited: 'Caption edited',
  bio_changed: 'Bio changed',
  followers_changed: 'Followers changed',
  avatar_changed: 'Avatar updated',
  display_name_changed: 'Display name changed',
}

function StatRow({ label, value }: { label: string; value: string | number | null | undefined }) {
  if (value == null) return null
  return (
    <div className="flex justify-between text-sm py-1 border-b border-border/40 last:border-0">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium">{value}</span>
    </div>
  )
}

export function SocialChangeDetailSheet({
  changeId,
  onClose,
}: Readonly<SocialChangeDetailSheetProps>) {
  const [detail, setDetail] = useState<SocialChangeDetail | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    if (!changeId) {
      setDetail(null)
      return
    }
    setIsLoading(true)
    SocialApi.getChange(changeId)
      .then(setDetail)
      .catch(() => setDetail(null))
      .finally(() => setIsLoading(false))
  }, [changeId])

  const change = detail?.change
  const before = detail?.beforeData
  const after = detail?.afterData

  return (
    <Sheet open={!!changeId} onOpenChange={(open) => !open && onClose()}>
      <SheetContent className="w-full sm:max-w-lg overflow-y-auto">
        <SheetHeader className="mb-4">
          <SheetTitle>Change detail</SheetTitle>
          {change && (
            <p className="text-xs text-muted-foreground">
              {formatRelativeTime(change.createdAt)}
            </p>
          )}
        </SheetHeader>

        {isLoading && (
          <div className="flex flex-col gap-3">
            <Skeleton className="h-6 w-full" />
            <Skeleton className="h-32 w-full" />
            <Skeleton className="h-32 w-full" />
          </div>
        )}

        {!isLoading && change && (
          <div className="flex flex-col gap-6">
            {/* Change types */}
            <div className="flex flex-wrap gap-2">
              {change.changeTypes.map((t) => (
                <Badge key={t} variant="outline" className="text-xs">
                  {CHANGE_TYPE_LABELS[t] ?? t}
                </Badge>
              ))}
            </div>

            {/* Followers delta */}
            {change.summary?.followersDelta !== undefined &&
              change.summary.followersDelta !== 0 && (
                <div>
                  <p className="text-xs text-muted-foreground uppercase tracking-wide mb-2">
                    Follower change
                  </p>
                  <Badge
                    variant={change.summary.followersDelta >= 0 ? 'default' : 'destructive'}
                    className="text-base font-semibold"
                  >
                    {change.summary.followersDelta >= 0 ? '+' : ''}
                    {change.summary.followersDelta}{' '}
                    {change.summary.followersDelta >= 0 ? '▲' : '▼'}
                  </Badge>
                </div>
              )}

            {/* Bio diff */}
            {change.summary?.bio && (
              <div>
                <p className="text-xs text-muted-foreground uppercase tracking-wide mb-2">Bio</p>
                {change.summary.bio.from && (
                  <p className="text-sm line-through text-muted-foreground bg-destructive/10 rounded px-2 py-1 mb-1">
                    {change.summary.bio.from}
                  </p>
                )}
                {change.summary.bio.to && (
                  <p className="text-sm bg-green-500/10 text-green-700 dark:text-green-400 rounded px-2 py-1">
                    {change.summary.bio.to}
                  </p>
                )}
              </div>
            )}

            {/* Display name diff */}
            {change.summary?.displayName && (
              <div>
                <p className="text-xs text-muted-foreground uppercase tracking-wide mb-2">
                  Display name
                </p>
                <p className="text-sm line-through text-muted-foreground bg-destructive/10 rounded px-2 py-1 mb-1">
                  {change.summary.displayName.from}
                </p>
                <p className="text-sm bg-green-500/10 text-green-700 dark:text-green-400 rounded px-2 py-1">
                  {change.summary.displayName.to}
                </p>
              </div>
            )}

            {/* Snapshot comparison */}
            {(before || after) && (
              <div className="grid grid-cols-2 gap-4">
                {before && (
                  <div>
                    <p className="text-xs text-muted-foreground uppercase tracking-wide mb-2">
                      Before
                    </p>
                    <div className="bg-muted/50 rounded p-3">
                      <StatRow label="Followers" value={before.followersCount.toLocaleString()} />
                      <StatRow label="Following" value={before.followingCount.toLocaleString()} />
                      <StatRow label="Posts" value={before.postsCount} />
                    </div>
                  </div>
                )}
                {after && (
                  <div>
                    <p className="text-xs text-muted-foreground uppercase tracking-wide mb-2">
                      After
                    </p>
                    <div className="bg-muted/50 rounded p-3">
                      <StatRow label="Followers" value={after.followersCount.toLocaleString()} />
                      <StatRow label="Following" value={after.followingCount.toLocaleString()} />
                      <StatRow label="Posts" value={after.postsCount} />
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        )}
      </SheetContent>
    </Sheet>
  )
}
