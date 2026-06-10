'use client'

import { Badge } from '@workspace/ui/components/atoms/badge'
import { formatDateTime } from '@workspace/ui'
import { useState } from 'react'
import type { SocialChange, SocialProfile } from '../../domain/types'
import { PlatformIcon } from '../platform-icon'
import { SocialChangeCard } from './social-change-card'

interface SocialChangesSectionProps {
  changes: SocialChange[]
  profiles: SocialProfile[]
  isLoading?: boolean
}

export function SocialChangesSection({
  changes,
  profiles,
  isLoading = false,
}: Readonly<SocialChangesSectionProps>) {
  const [selectedChangeId, setSelectedChangeId] = useState<string | null>(null)

  const profileMap = new Map<string, SocialProfile>(profiles.map((p) => [p.id, p]))

  const selectedChange = selectedChangeId
    ? changes.find((c) => c.id === selectedChangeId) ?? null
    : null

  if (isLoading) {
    return (
      <div className="flex flex-col gap-2 mt-8">
        <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide">
          Social Activity
        </h2>
        <div className="flex flex-col gap-2">
          {[1, 2].map((i) => (
            <div key={i} className="h-10 bg-muted animate-pulse rounded-md" />
          ))}
        </div>
      </div>
    )
  }

  if (changes.length === 0) return null

  return (
    <div className="flex flex-col gap-4 mt-8 border-t border-border pt-6">
      <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide">
        Social Activity
      </h2>

      <div className="grid grid-cols-1 md:grid-cols-[280px_1fr] gap-4">
        {/* Feed list */}
        <div className="flex flex-col gap-1">
          {changes.map((change) => {
            const profile = profileMap.get(change.profileId)
            const isSelected = selectedChangeId === change.id
            return (
              <button
                key={change.id}
                type="button"
                onClick={() => setSelectedChangeId(isSelected ? null : change.id)}
                className={[
                  'flex items-center gap-2 px-3 py-2 rounded-md text-left transition-colors',
                  isSelected
                    ? 'bg-primary/10 text-foreground'
                    : 'hover:bg-muted/60 text-muted-foreground hover:text-foreground',
                ].join(' ')}
              >
                {profile && (
                  <PlatformIcon platform={profile.platform} className="w-3.5 h-3.5 shrink-0" />
                )}
                <span className="text-sm font-medium truncate flex-1">
                  {profile ? `@${profile.handle}` : change.profileId.slice(0, 8)}
                </span>
                <Badge variant="secondary" className="text-xs shrink-0 px-1.5 py-0">
                  Social
                </Badge>
                <span className="text-xs text-muted-foreground shrink-0">
                  {formatDateTime(change.createdAt)}
                </span>
              </button>
            )
          })}
        </div>

        {/* Detail panel */}
        <div>
          {selectedChange ? (
            <SocialChangeCard change={selectedChange} />
          ) : (
            <div className="flex items-center justify-center h-24 rounded-lg border border-border border-dashed text-sm text-muted-foreground">
              Select a social change to see details
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
