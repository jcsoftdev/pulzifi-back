import { Badge } from '@workspace/ui/components/atoms/badge'
import { Card, CardContent } from '@workspace/ui/components/atoms/card'
import { Avatar, AvatarFallback, AvatarImage } from '@workspace/ui/components/atoms/avatar'
import { formatRelativeTime } from '@workspace/ui/lib/date'
import { cn } from '@workspace/ui/lib/utils'
import type { SocialProfile } from '../domain/types'
import { PLATFORM_LABELS } from '../domain/types'
import { PlatformIcon } from './platform-icon'

interface SocialProfileCardProps {
  profile: SocialProfile
  onClick?: () => void
  className?: string
}

export function SocialProfileCard({
  profile,
  onClick,
  className,
}: Readonly<SocialProfileCardProps>) {
  const timeSince = profile.lastCheckedAt
    ? formatRelativeTime(profile.lastCheckedAt)
    : 'Never checked'

  const initials = (profile.displayName || profile.handle)
    .slice(0, 2)
    .toUpperCase()

  return (
    <Card
      className={cn(
        'cursor-pointer hover:border-primary/50 transition-colors',
        !profile.isActive && 'opacity-60',
        className
      )}
      onClick={onClick}
    >
      <CardContent className="p-4 flex items-start gap-3">
        <Avatar className="h-10 w-10 shrink-0">
          {profile.avatarUrl ? (
            <AvatarImage src={profile.avatarUrl} alt={profile.handle} />
          ) : null}
          <AvatarFallback className="text-xs">{initials}</AvatarFallback>
        </Avatar>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <span className="font-medium text-sm truncate">@{profile.handle}</span>
            <Badge variant="outline" className="flex items-center gap-1 text-xs shrink-0">
              <PlatformIcon platform={profile.platform} className="w-3 h-3" />
              {PLATFORM_LABELS[profile.platform] ?? profile.platform}
            </Badge>
            <Badge variant="secondary" className="text-xs shrink-0">
              Social
            </Badge>
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
