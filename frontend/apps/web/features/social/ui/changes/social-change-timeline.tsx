import { Skeleton } from '@workspace/ui/components/atoms/skeleton'
import type { SocialChange } from '../../domain/types'
import { SocialChangeCard } from './social-change-card'

interface SocialChangeTimelineProps {
  changes: SocialChange[]
  isLoading?: boolean
}

export function SocialChangeTimeline({
  changes,
  isLoading = false,
}: Readonly<SocialChangeTimelineProps>) {
  if (isLoading) {
    return (
      <div className="flex flex-col gap-4">
        {[1, 2, 3].map((i) => (
          <Skeleton key={i} className="h-24 w-full rounded-lg" />
        ))}
      </div>
    )
  }

  if (changes.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-8 text-center text-muted-foreground">
        <p className="text-sm">No changes detected yet.</p>
        <p className="text-xs mt-1">Changes will appear here after the first check.</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      {changes.map((change) => (
        <SocialChangeCard key={change.id} change={change} />
      ))}
    </div>
  )
}
