import { Badge } from '@workspace/ui/components/atoms/badge'
import { cn } from '@workspace/ui/lib/utils'
import type { SocialQuotaStatus } from '../domain/types'

interface SocialQuotaBadgeProps {
  quota: SocialQuotaStatus | null
  className?: string
}

export function SocialQuotaBadge({ quota, className }: Readonly<SocialQuotaBadgeProps>) {
  if (!quota) {
    return (
      <Badge variant="secondary" className={cn('text-xs', className)}>
        Loading quota...
      </Badge>
    )
  }

  const isUnlimited = quota.checksLimit === null
  const isExhausted = !isUnlimited && quota.checksUsed >= (quota.checksLimit ?? 0)

  const label = isUnlimited
    ? `${quota.checksUsed} checks today (unlimited)`
    : `${quota.checksUsed}/${quota.checksLimit} checks today`

  return (
    <Badge
      variant={isExhausted ? 'destructive' : 'secondary'}
      className={cn('text-xs', className)}
    >
      {label}
    </Badge>
  )
}
