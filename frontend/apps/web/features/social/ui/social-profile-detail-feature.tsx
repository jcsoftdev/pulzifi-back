'use client'

import { Avatar, AvatarFallback, AvatarImage } from '@workspace/ui/components/atoms/avatar'
import { Badge } from '@workspace/ui/components/atoms/badge'
import { Button } from '@workspace/ui/components/atoms/button'
import { Card, CardContent } from '@workspace/ui/components/atoms/card'
import { Heart, ImageOff, Loader2, MessageCircle, RefreshCw, Trophy } from 'lucide-react'
import { useState } from 'react'
import type { SocialPost } from '../domain/types'
import { useSocialChanges } from '../application/use-social-changes'
import { useSocialSnapshots } from '../application/use-social-snapshots'
import type { SocialChange, SocialProfile, SocialSnapshot } from '../domain/types'
import type { SocialSnapshotSummary } from '@workspace/services/social-api'
import { PLATFORM_LABELS } from '../domain/types'
import { SocialChangeDetailSheet } from './changes/social-change-detail-sheet'
import { PlatformIcon } from './platform-icon'
import { useSocialProfileDetail } from '../application/use-social-profile-detail'
import { SocialSnapshotsHistory } from './social-snapshots-history'

interface SocialProfileDetailFeatureProps {
  workspaceId: string
  profileId: string
  initialProfile: SocialProfile
  initialSnapshot: SocialSnapshot | null
  initialChanges: SocialChange[]
  initialSnapshots: SocialSnapshotSummary[]
  initialPrevSnapshot: SocialSnapshot | null
}

function DeltaBadge({ delta }: { delta: number | null }) {
  if (delta === null || delta === 0) return null
  const up = delta > 0
  return (
    <span className={up ? 'text-green-400' : 'text-red-400'}>
      {up ? '↑' : '↓'}
      {Math.abs(delta).toLocaleString()}
    </span>
  )
}

function PostThumbnail({
  src,
  alt,
  likesCount,
  commentsCount,
  likesDelta,
  commentsDelta,
}: {
  src: string
  alt: string
  likesCount: number
  commentsCount: number
  likesDelta: number | null
  commentsDelta: number | null
}) {
  const [failed, setFailed] = useState(false)
  const [hovered, setHovered] = useState(false)

  if (failed) {
    return (
      <div className="w-full h-full flex flex-col items-center justify-center gap-1 text-muted-foreground/40">
        <ImageOff className="w-6 h-6" />
      </div>
    )
  }

  return (
    // biome-ignore lint/a11y/noStaticElementInteractions: hover overlay is visual-only enhancement
    <div
      className="relative w-full h-full"
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      {/* biome-ignore lint/performance/noImgElement: post media URLs are dynamic external URLs with unknown dimensions */}
      <img
        src={src}
        alt={alt}
        className="w-full h-full object-cover"
        onError={() => setFailed(true)}
      />
      {hovered && (likesCount > 0 || commentsCount > 0) && (
        <div className="absolute inset-0 bg-black/50 flex items-center justify-center gap-3 text-white text-sm font-medium">
          <span className="flex items-center gap-1">
            <Heart className="w-4 h-4 fill-white" />
            {likesCount.toLocaleString()}
            <DeltaBadge delta={likesDelta} />
          </span>
          <span className="flex items-center gap-1">
            <MessageCircle className="w-4 h-4 fill-white" />
            {commentsCount.toLocaleString()}
            <DeltaBadge delta={commentsDelta} />
          </span>
        </div>
      )}
    </div>
  )
}

function PostsGrid({
  snapshot,
  prevPosts,
}: {
  snapshot: SocialSnapshot | null
  prevPosts: SocialPost[]
}) {
  if (!snapshot?.data?.posts?.length) {
    return (
      <p className="text-sm text-muted-foreground">No posts captured yet.</p>
    )
  }

  // No previous snapshot → show current counts only (no deltas).
  const prevByExternalId = new Map(prevPosts.map((p) => [p.externalId, p]))

  return (
    <div className="grid grid-cols-3 gap-2">
      {snapshot.data.posts.map((post, i) => {
        const src = post.storedMediaUrl || post.mediaUrl
        const likes = post.likesCount ?? 0
        const comments = post.commentsCount ?? 0
        const prev = prevByExternalId.get(post.externalId)
        const likesDelta = prev ? likes - (prev.likesCount ?? 0) : null
        const commentsDelta = prev ? comments - (prev.commentsCount ?? 0) : null
        return (
          <div key={post.externalId || i} className="aspect-square bg-muted rounded overflow-hidden">
            {src ? (
              <PostThumbnail
                src={src}
                alt={post.caption || 'Post'}
                likesCount={likes}
                commentsCount={comments}
                likesDelta={likesDelta}
                commentsDelta={commentsDelta}
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center text-muted-foreground/40">
                <ImageOff className="w-6 h-6" />
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}

function fmt(n: number) {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`
  return n.toLocaleString()
}

function EngagementRanking({ posts, prevPosts }: { posts: SocialPost[]; prevPosts: SocialPost[] }) {
  const ranked = [...posts]
    .filter((p) => p.likesCount > 0 || p.commentsCount > 0)
    .sort((a, b) => (b.likesCount + b.commentsCount) - (a.likesCount + a.commentsCount))
    .slice(0, 5)

  if (ranked.length === 0) {
    return <p className="text-sm text-muted-foreground">No engagement data yet.</p>
  }

  const topTotal = (ranked[0]?.likesCount ?? 0) + (ranked[0]?.commentsCount ?? 0)
  const prevByExternalId = new Map(prevPosts.map((p) => [p.externalId, p]))

  return (
    <div className="flex flex-col gap-2">
      {ranked.map((post, i) => {
        const total = post.likesCount + post.commentsCount
        const pct = topTotal > 0 ? (total / topTotal) * 100 : 0
        const src = post.storedMediaUrl || post.mediaUrl
        const prev = prevByExternalId.get(post.externalId)
        const likesDelta = prev ? post.likesCount - prev.likesCount : null
        const commentsDelta = prev ? post.commentsCount - prev.commentsCount : null
        return (
          <div key={post.externalId || i} className="flex items-center gap-3">
            <span className="w-5 text-xs text-muted-foreground text-right shrink-0">
              {i === 0 ? <Trophy className="w-4 h-4 text-yellow-500" /> : `#${i + 1}`}
            </span>
            <div className="w-10 h-10 shrink-0 rounded overflow-hidden bg-muted">
              {src ? (
                // biome-ignore lint/performance/noImgElement: ranking thumbnail URLs are dynamic external URLs
                <img src={src} alt={post.caption || 'Post'} className="w-full h-full object-cover" />
              ) : (
                <div className="w-full h-full flex items-center justify-center text-muted-foreground/30">
                  <ImageOff className="w-4 h-4" />
                </div>
              )}
            </div>
            <div className="flex-1 min-w-0">
              <div className="h-1.5 rounded-full bg-muted overflow-hidden mb-1">
                <div
                  className="h-full rounded-full bg-primary transition-all"
                  style={{ width: `${pct}%` }}
                />
              </div>
              <p className="text-xs text-muted-foreground truncate">
                {post.caption || '—'}
              </p>
            </div>
            <div className="flex items-center gap-2 shrink-0 text-xs text-muted-foreground">
              <span className="flex items-center gap-0.5">
                <Heart className="w-3 h-3" />
                {fmt(post.likesCount)}
                {likesDelta !== null && likesDelta !== 0 && (
                  <span className={likesDelta > 0 ? 'text-green-500' : 'text-red-500'}>
                    {likesDelta > 0 ? '+' : ''}{fmt(Math.abs(likesDelta))}
                  </span>
                )}
              </span>
              <span className="flex items-center gap-0.5">
                <MessageCircle className="w-3 h-3" />
                {fmt(post.commentsCount)}
                {commentsDelta !== null && commentsDelta !== 0 && (
                  <span className={commentsDelta > 0 ? 'text-green-500' : 'text-red-500'}>
                    {commentsDelta > 0 ? '+' : ''}{fmt(Math.abs(commentsDelta))}
                  </span>
                )}
              </span>
            </div>
          </div>
        )
      })}
    </div>
  )
}

function ProfileHeader({
  profile,
  snapshot,
  isChecking,
  checkError,
  onRunCheck,
}: {
  profile: SocialProfile
  snapshot: SocialSnapshot | null
  isChecking: boolean
  checkError: string | null
  onRunCheck: () => void
}) {
  const initials = (profile.displayName || profile.handle).slice(0, 2).toUpperCase()

  return (
    <Card>
      <CardContent className="p-6 flex flex-col sm:flex-row items-start sm:items-center gap-4">
        <Avatar className="h-16 w-16 shrink-0">
          {profile.avatarUrl ? (
            <AvatarImage src={profile.avatarUrl} alt={profile.handle} />
          ) : null}
          <AvatarFallback className="text-lg">{initials}</AvatarFallback>
        </Avatar>

        <div className="flex-1 min-w-0">
          <div className="flex flex-wrap items-center gap-2 mb-1">
            <h1 className="text-xl font-semibold">@{profile.handle}</h1>
            <Badge variant="outline" className="flex items-center gap-1 text-xs">
              <PlatformIcon platform={profile.platform} className="w-3 h-3" />
              {PLATFORM_LABELS[profile.platform] ?? profile.platform}
            </Badge>
            <Badge variant="secondary" className="text-xs">
              Social
            </Badge>
            {!profile.isActive && (
              <Badge variant="destructive" className="text-xs">
                Inactive
              </Badge>
            )}
          </div>

          {profile.displayName && (
            <p className="text-sm text-muted-foreground mb-1">{profile.displayName}</p>
          )}

          {snapshot?.data?.bio && (
            <p className="text-sm text-muted-foreground mb-1">{snapshot.data.bio}</p>
          )}

          {snapshot != null && (
            <p className="text-sm text-muted-foreground">
              {(snapshot.followersCount ?? snapshot.data?.followersCount ?? 0).toLocaleString()} followers
            </p>
          )}

          {checkError && (
            <p className="text-xs text-destructive mt-1">{checkError}</p>
          )}
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={onRunCheck}
          disabled={isChecking}
          className="shrink-0"
        >
          {isChecking ? (
            <Loader2 className="w-4 h-4 animate-spin mr-2" />
          ) : (
            <RefreshCw className="w-4 h-4 mr-2" />
          )}
          {isChecking ? 'Checking…' : 'Run Check'}
        </Button>
      </CardContent>
    </Card>
  )
}

export function SocialProfileDetailFeature({
  workspaceId: _workspaceId,
  profileId,
  initialProfile,
  initialSnapshot,
  initialChanges,
  initialSnapshots,
  initialPrevSnapshot,
}: Readonly<SocialProfileDetailFeatureProps>) {
  const { profile, snapshot, isChecking, checkError, runCheck } =
    useSocialProfileDetail(profileId, initialProfile, initialSnapshot)
  const { changes, fetchChanges } = useSocialChanges(profileId, initialChanges)
  const { snapshots, fetchSnapshots } = useSocialSnapshots(profileId, initialSnapshots)

  const [selectedChangeId, setSelectedChangeId] = useState<string | null>(null)
  const handleRunCheck = () => runCheck(() => { fetchChanges(); fetchSnapshots() })

  // Map each snapshot to the change it triggered (change.toSnapshotId -> change.id)
  const changesBySnapshotId = new Map(changes.map((c) => [c.toSnapshotId, c.id]))

  return (
    <div className="flex-1 px-4 md:px-8 py-6 max-w-4xl mx-auto w-full flex flex-col gap-6">
      <ProfileHeader
        profile={profile}
        snapshot={snapshot}
        isChecking={isChecking}
        checkError={checkError}
        onRunCheck={handleRunCheck}
      />

      <section>
        <h2 className="text-base font-medium mb-3">Latest Posts</h2>
        <PostsGrid snapshot={snapshot} prevPosts={initialPrevSnapshot?.data?.posts ?? []} />
      </section>

      {snapshot?.data?.posts && snapshot.data.posts.length > 0 && (
        <section>
          <h2 className="text-base font-medium mb-3">Top Posts by Engagement</h2>
          <EngagementRanking
            posts={snapshot.data.posts}
            prevPosts={initialPrevSnapshot?.data?.posts ?? []}
          />
        </section>
      )}

      <section>
        <h2 className="text-base font-medium mb-3">Run History</h2>
        <SocialSnapshotsHistory
          snapshots={snapshots}
          isLoading={false}
          changesBySnapshotId={changesBySnapshotId}
          onChangeClick={setSelectedChangeId}
        />
      </section>

      <SocialChangeDetailSheet
        changeId={selectedChangeId}
        onClose={() => setSelectedChangeId(null)}
      />
    </div>
  )
}
