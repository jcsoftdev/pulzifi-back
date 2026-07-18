'use client'

import { Avatar, AvatarFallback, AvatarImage } from '@workspace/ui/components/atoms/avatar'
import { Badge } from '@workspace/ui/components/atoms/badge'
import { Card, CardContent } from '@workspace/ui/components/atoms/card'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@workspace/ui/components/molecules/dropdown-menu'
import { formatRelativeTime } from '@workspace/ui/lib/date'
import { cn } from '@workspace/ui/lib/utils'
import { MoreHorizontal, Trash2 } from 'lucide-react'
import { useState } from 'react'
import type { SocialProfile } from '../domain/types'
import { PLATFORM_LABELS } from '../domain/types'
import { PlatformIcon } from './platform-icon'

interface SocialProfileCardProps {
  profile: SocialProfile
  className?: string
  onDelete?: (profile: SocialProfile) => void
}

export function SocialProfileCard({
  profile,
  className,
  onDelete,
}: Readonly<SocialProfileCardProps>) {
  const [menuOpen, setMenuOpen] = useState(false)

  const timeSince = profile.lastCheckedAt
    ? formatRelativeTime(profile.lastCheckedAt)
    : 'Never checked'

  const initials = (profile.displayName || profile.handle)
    .slice(0, 2)
    .toUpperCase()

  return (
    <Card
      className={cn(
        'group relative cursor-pointer hover:border-primary/50 transition-colors',
        !profile.isActive && 'opacity-60',
        className
      )}
    >
      {onDelete && (
        <div className="absolute top-2 right-2 z-10">
          <DropdownMenu open={menuOpen} onOpenChange={setMenuOpen}>
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                // The card is wrapped in a <Link>; stop the click from navigating.
                onClick={(e) => {
                  e.preventDefault()
                  e.stopPropagation()
                }}
                className="p-1 rounded-md text-muted-foreground opacity-0 group-hover:opacity-100 focus:opacity-100 hover:bg-muted transition-opacity"
                aria-label={`Actions for @${profile.handle}`}
              >
                <MoreHorizontal className="w-4 h-4" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-40">
              <DropdownMenuItem
                variant="destructive"
                onClick={(e) => {
                  e.preventDefault()
                  e.stopPropagation()
                  onDelete(profile)
                }}
              >
                <Trash2 className="w-4 h-4" />
                Delete profile
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )}

      <CardContent className="p-4 flex items-start gap-3">
        <Avatar className="h-10 w-10 shrink-0">
          {profile.avatarUrl ? (
            <AvatarImage src={profile.avatarUrl} alt={profile.handle} />
          ) : null}
          <AvatarFallback className="text-xs">{initials}</AvatarFallback>
        </Avatar>

        <div className="flex-1 min-w-0">
          <p className="font-medium text-sm truncate mb-1">@{profile.handle}</p>

          <div className="flex items-center flex-wrap gap-1 mb-1">
            <Badge variant="outline" className="flex items-center gap-1 text-xs">
              <PlatformIcon platform={profile.platform} className="w-3 h-3" />
              {PLATFORM_LABELS[profile.platform] ?? profile.platform}
            </Badge>
            {!profile.isActive && (
              <Badge variant="destructive" className="text-xs">Inactive</Badge>
            )}
          </div>

          {profile.displayName && (
            <p className="text-xs text-muted-foreground truncate mb-1">{profile.displayName}</p>
          )}

          <p className="text-xs text-muted-foreground">{timeSince}</p>
        </div>
      </CardContent>
    </Card>
  )
}
