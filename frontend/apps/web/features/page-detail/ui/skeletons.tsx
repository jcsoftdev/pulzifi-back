import { Skeleton } from '@workspace/ui'

export function PageInfoSkeleton() {
  return (
    <div className="flex flex-col gap-6 bg-card p-6 rounded-xl shadow-sm border border-border">
      <div className="flex flex-col md:flex-row justify-between gap-6">
        <div className="flex gap-4">
          <div className="flex flex-col gap-2 w-full">
            <div className="flex items-center gap-2">
              <Skeleton className="h-8 w-48" />
              <Skeleton className="h-6 w-6 rounded-md" />
            </div>
            <Skeleton className="h-5 w-64" />
          </div>
        </div>

        <div className="flex items-center gap-3">
          <Skeleton className="h-10 w-24" />
          <Skeleton className="h-10 w-32" />
        </div>
      </div>
    </div>
  )
}

export function ChecksHistorySkeleton() {
  return (
    <div className="flex flex-col gap-6 bg-card border border-border rounded-xl p-6 h-full min-h-[500px]">
      <div className="flex items-center justify-between">
        <Skeleton className="h-7 w-40" />
      </div>

      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-2">
          <Skeleton className="h-5 w-16" />
          <div className="h-[1px] flex-1 bg-border" />
        </div>

        <div className="space-y-8 pl-2">
          {[
            1,
            2,
            3,
            4,
            5,
          ].map((i) => (
            <div key={i} className="flex flex-col gap-2 pl-6 relative">
              <Skeleton className="absolute left-0 top-1.5 h-3 w-3 rounded-full" />
              <div className="flex flex-col gap-1">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-4 w-32" />
                {i % 2 === 0 && <Skeleton className="mt-2 h-7 w-40 rounded-md" />}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export function AdvancedSettingsSkeleton() {
  return (
    <div className="flex flex-col gap-6 bg-card border border-border rounded-xl p-6 h-full">
      <Skeleton className="h-7 w-48" /> {/* Title */}
      <div className="flex flex-col gap-4">
        {/* Legend Row */}
        <div className="flex flex-wrap items-center gap-3">
          {[
            1,
            2,
            3,
          ].map((i) => (
            <div key={i} className="flex items-center gap-2">
              <Skeleton className="h-4 w-4 rounded-full" />
              <Skeleton className="h-4 w-20" />
            </div>
          ))}
        </div>
        <Skeleton className="h-3 w-64" /> {/* Helper text */}
        {/* Badges Row */}
        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-28 rounded-md" />
          <Skeleton className="h-8 w-32 rounded-md" />
        </div>
        <Skeleton className="h-3 w-56" /> {/* Helper text */}
      </div>
    </div>
  )
}

export function GeneralSummarySkeleton() {
  return (
    <div className="flex flex-col gap-6 bg-card border border-border rounded-xl p-6 h-full">
      <Skeleton className="h-7 w-48" /> {/* Title */}
      <div className="flex flex-col gap-4">
        {/* Tag Section */}
        <div className="flex flex-col gap-2">
          <Skeleton className="h-4 w-10" /> {/* Label */}
          <div className="flex items-center gap-2">
            <Skeleton className="h-8 w-24 rounded-md" />
            <Skeleton className="h-4 w-4" /> {/* Pencil */}
          </div>
        </div>

        {/* Check Frequency */}
        <div className="flex flex-col gap-2">
          <Skeleton className="h-4 w-32" /> {/* Label */}
          <Skeleton className="h-10 w-full rounded-md" /> {/* Dropdown */}
        </div>

        {/* Options List */}
        <div className="flex flex-col gap-3 mt-2">
          <div className="flex items-center gap-2">
            <Skeleton className="h-4 w-4 rounded-sm" />
            <Skeleton className="h-4 w-48" />
          </div>
        </div>
      </div>
    </div>
  )
}

export function IntelligentInsightsSkeleton() {
  return (
    <div className="flex flex-col gap-6 bg-card border border-border rounded-xl p-6 h-full">
      <Skeleton className="h-7 w-48" /> {/* Title */}
      <div className="flex flex-col gap-3">
        <div className="flex items-center gap-2">
          <Skeleton className="h-4 w-4 rounded-sm" />
          <Skeleton className="h-4 w-32" />
        </div>
      </div>
    </div>
  )
}
