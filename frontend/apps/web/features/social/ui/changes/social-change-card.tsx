import { Badge } from '@workspace/ui/components/atoms/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@workspace/ui/components/atoms/card'
import { formatRelativeTime } from '@workspace/ui/lib/date'
import type { ChangeType, SocialChange } from '../../domain/types'

// ─── Per-change-type renderers ───────────────────────────────────────────────

function FollowersChangedRenderer({ delta }: { delta: number }) {
  const isPositive = delta >= 0
  return (
    <div className="flex items-center gap-2">
      <Badge variant={isPositive ? 'default' : 'destructive'} className="text-base font-semibold">
        {isPositive ? '+' : ''}
        {delta} {isPositive ? '▲' : '▼'}
      </Badge>
      <span className="text-sm text-muted-foreground">followers change</span>
    </div>
  )
}

function BioChangedRenderer({ from, to }: { from: string; to: string }) {
  return (
    <div className="flex flex-col gap-1 text-sm">
      <p className="text-muted-foreground text-xs uppercase tracking-wide">Bio changed</p>
      {from && (
        <p className="line-through text-muted-foreground bg-destructive/10 rounded px-2 py-1">
          {from}
        </p>
      )}
      {to && (
        <p className="bg-green-500/10 text-green-700 dark:text-green-400 rounded px-2 py-1">
          {to}
        </p>
      )}
    </div>
  )
}

function NewPostRenderer({
  caption,
  storedMediaUrl,
  postedAt,
}: {
  caption: string
  storedMediaUrl: string
  postedAt: string
}) {
  const timeSince = formatRelativeTime(postedAt)
  return (
    <div className="flex gap-3">
      {storedMediaUrl && (
        // biome-ignore lint/performance/noImgElement: stored media URL is dynamic external URL
        <img
          src={storedMediaUrl}
          alt="Post thumbnail"
          className="h-16 w-16 object-cover rounded shrink-0"
        />
      )}
      <div className="flex flex-col gap-1 min-w-0">
        <p className="text-sm line-clamp-2">{caption || 'No caption'}</p>
        <p className="text-xs text-muted-foreground">{timeSince}</p>
      </div>
    </div>
  )
}

function GenericRenderer({ types }: { types: ChangeType[] }) {
  const labels: Record<string, string> = {
    removed_post: 'Post removed',
    caption_edited: 'Caption edited',
    avatar_changed: 'Avatar updated',
    display_name_changed: 'Display name changed',
  }
  return (
    <div className="flex flex-wrap gap-1">
      {types.map((t) => (
        <Badge key={t} variant="outline" className="text-xs">
          {labels[t] ?? t}
        </Badge>
      ))}
    </div>
  )
}

// ─── Main component ──────────────────────────────────────────────────────────

interface SocialChangeCardProps {
  change: SocialChange
  onClick?: (changeId: string) => void
}

export function SocialChangeCard({ change, onClick }: Readonly<SocialChangeCardProps>) {
  const timeSince = formatRelativeTime(change.createdAt)
  const summary = change.summary

  const hasFollowers = change.changeTypes.includes('followers_changed')
  const hasBio = change.changeTypes.includes('bio_changed')
  const hasNewPost = change.changeTypes.includes('new_post')

  const otherTypes = change.changeTypes.filter(
    (t) => t !== 'followers_changed' && t !== 'bio_changed' && t !== 'new_post'
  )

  return (
    <Card
      className={onClick ? 'cursor-pointer hover:border-primary/50 transition-colors' : undefined}
      onClick={onClick ? () => onClick(change.id) : undefined}
    >
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium flex items-center justify-between gap-2">
          <span>Change detected</span>
          <span className="text-xs font-normal text-muted-foreground">{timeSince}</span>
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        {hasFollowers && summary?.followersDelta !== undefined && (
          <FollowersChangedRenderer delta={summary.followersDelta} />
        )}

        {hasBio && summary?.bio && (
          <BioChangedRenderer from={summary.bio.from} to={summary.bio.to} />
        )}

        {hasNewPost &&
          (summary?.newPosts ?? []).length > 0 && (
            <div className="flex flex-col gap-2">
              <p className="text-xs text-muted-foreground uppercase tracking-wide">
                New post{(summary?.newPosts?.length ?? 0) > 1 ? 's' : ''}
              </p>
              <NewPostRenderer
                caption=""
                storedMediaUrl=""
                postedAt={change.createdAt}
              />
            </div>
          )}

        {otherTypes.length > 0 && <GenericRenderer types={otherTypes} />}
      </CardContent>
    </Card>
  )
}
