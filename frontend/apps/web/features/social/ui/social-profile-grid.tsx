'use client'

import { useRouter } from 'next/navigation'
import type { SocialProfile } from '../domain/types'
import { SocialProfileCard } from './social-profile-card'

interface SocialProfileGridProps {
  profiles: SocialProfile[]
  workspaceId: string
}

export function SocialProfileGrid({
  profiles,
  workspaceId,
}: Readonly<SocialProfileGridProps>) {
  const router = useRouter()

  if (profiles.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center text-muted-foreground">
        <p className="text-sm">No social profiles added yet.</p>
        <p className="text-xs mt-1">Add a profile to start monitoring.</p>
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {profiles.map((profile) => (
        <SocialProfileCard
          key={profile.id}
          profile={profile}
          onClick={() => router.push(`/workspaces/${workspaceId}/social/${profile.id}`)}
        />
      ))}
    </div>
  )
}
