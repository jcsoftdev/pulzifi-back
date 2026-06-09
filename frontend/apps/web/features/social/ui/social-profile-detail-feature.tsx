'use client'

import { Avatar, AvatarFallback, AvatarImage } from '@workspace/ui/components/atoms/avatar'
import { Badge } from '@workspace/ui/components/atoms/badge'
import { Card, CardContent } from '@workspace/ui/components/atoms/card'
import { Skeleton } from '@workspace/ui/components/atoms/skeleton'
import { useEffect } from 'react'
import { useSocialChanges } from '../application/use-social-changes'
import type { SocialProfile, SocialSnapshot } from '../domain/types'
import { PLATFORM_LABELS } from '../domain/types'
import { SocialChangeTimeline } from './changes/social-change-timeline'
import { PlatformIcon } from './platform-icon'
import { useSocialProfileDetail } from '../application/use-social-profile-detail'

interface SocialProfileDetailFeatureProps {
  workspaceId: string
  profileId: string
}

function PostsGrid({ snapshot }: { snapshot: SocialSnapshot | null }) {
  if (!snapshot?.data?.posts?.length) {
    return (
      <p className="text-sm text-muted-foreground">No posts captured yet.</p>
    )
  }

  return (
    <div className="grid grid-cols-3 gap-2">
      {snapshot.data.posts.map((post) => (
        <div key={post.externalId} className="aspect-square bg-muted rounded overflow-hidden">
          {post.storedMediaUrl ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={post.storedMediaUrl}
              alt={post.caption || 'Post'}
              className="w-full h-full object-cover"
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center text-xs text-muted-foreground">
              No image
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

function ProfileHeader({
  profile,
  snapshot,
}: {
  profile: SocialProfile
  snapshot: SocialSnapshot | null
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

          {snapshot?.data && (
            <p className="text-sm text-muted-foreground">
              {snapshot.data.followersCount.toLocaleString()} followers
            </p>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

export function SocialProfileDetailFeature({
  workspaceId: _workspaceId,
  profileId,
}: Readonly<SocialProfileDetailFeatureProps>) {
  const { profile, snapshot, isLoading: profileLoading, fetchProfile } =
    useSocialProfileDetail(profileId)
  const { changes, isLoading: changesLoading, fetchChanges } = useSocialChanges(profileId)

  useEffect(() => {
    fetchProfile()
    fetchChanges()
  }, [fetchProfile, fetchChanges])

  if (profileLoading) {
    return (
      <div className="flex-1 px-4 md:px-8 py-6 max-w-4xl mx-auto w-full flex flex-col gap-6">
        <Skeleton className="h-32 w-full rounded-lg" />
        <Skeleton className="h-48 w-full rounded-lg" />
        <Skeleton className="h-64 w-full rounded-lg" />
      </div>
    )
  }

  if (!profile) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <p className="text-muted-foreground">Profile not found.</p>
      </div>
    )
  }

  return (
    <div className="flex-1 px-4 md:px-8 py-6 max-w-4xl mx-auto w-full flex flex-col gap-6">
      <ProfileHeader profile={profile} snapshot={snapshot} />

      <section>
        <h2 className="text-base font-medium mb-3">Latest Posts</h2>
        <PostsGrid snapshot={snapshot} />
      </section>

      <section>
        <h2 className="text-base font-medium mb-3">Change Timeline</h2>
        <SocialChangeTimeline changes={changes} isLoading={changesLoading} />
      </section>
    </div>
  )
}
