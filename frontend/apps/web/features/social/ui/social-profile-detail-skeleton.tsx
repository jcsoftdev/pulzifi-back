import { Card, CardContent } from '@workspace/ui/components/atoms/card'
import { Skeleton } from '@workspace/ui/components/atoms/skeleton'

export function SocialProfileDetailSkeleton() {
  return (
    <div className="flex-1 px-4 md:px-8 py-6 max-w-4xl mx-auto w-full flex flex-col gap-6">
      {/* Profile header */}
      <Card>
        <CardContent className="p-6 flex flex-col sm:flex-row items-start sm:items-center gap-4">
          <Skeleton className="h-16 w-16 rounded-full shrink-0" />
          <div className="flex-1 min-w-0 space-y-2">
            <div className="flex items-center gap-2">
              <Skeleton className="h-6 w-32" />
              <Skeleton className="h-5 w-20" />
              <Skeleton className="h-5 w-14" />
            </div>
            <Skeleton className="h-4 w-64" />
            <Skeleton className="h-4 w-24" />
          </div>
          <Skeleton className="h-9 w-28 shrink-0" />
        </CardContent>
      </Card>

      {/* Latest Posts */}
      <section>
        <Skeleton className="h-5 w-28 mb-3" />
        <div className="grid grid-cols-3 gap-2">
          {Array.from({ length: 9 }).map((_, i) => (
            // biome-ignore lint/suspicious/noArrayIndexKey: static skeleton
            <Skeleton key={i} className="aspect-square rounded" />
          ))}
        </div>
      </section>

      {/* Run History */}
      <section>
        <Skeleton className="h-5 w-24 mb-3" />
        <div className="relative border-l border-border ml-2 space-y-6">
          {Array.from({ length: 4 }).map((_, i) => (
            // biome-ignore lint/suspicious/noArrayIndexKey: static skeleton
            <div key={i} className="relative pl-6">
              <div className="absolute -left-1.5 top-1.5 h-3 w-3 rounded-full bg-muted border-2 border-background" />
              <div className="space-y-1">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-4 w-48" />
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  )
}
