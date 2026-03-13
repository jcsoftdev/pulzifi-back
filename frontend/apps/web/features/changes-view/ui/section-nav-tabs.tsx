'use client'

import type { Check, MonitoredSection } from '@workspace/services/page-api'
import { cn } from '@workspace/ui/lib/utils'
import { getSectionColor } from '@/features/page/domain/section-colors'

interface SectionNavTabsProps {
  sections: MonitoredSection[]
  selectedSectionId: string
  onSelect: (id: string) => void
  checks: Check[]
  activeSectionChecks?: Check[]
}

export function SectionNavTabs({
  sections,
  selectedSectionId,
  onSelect,
  checks,
  activeSectionChecks = [],
}: Readonly<SectionNavTabsProps>) {
  // Count actual content diff items (text blocks that changed) per section.
  const diffCountBySection = new Map<string, number>()
  let totalDiffCount = 0
  for (const sc of activeSectionChecks) {
    const count = sc.contentDiff?.total_changes ?? 0
    if (sc.sectionId) diffCountBySection.set(sc.sectionId, count)
    totalDiffCount += count
  }

  return (
    <div className="overflow-x-auto pb-1 px-4 md:px-0 -mx-4 md:mx-0">
      <div className="flex items-center gap-1.5 min-w-max px-4 md:px-0">
        {/* All tab */}
        <button
          type="button"
          onClick={() => onSelect('all')}
          style={{ touchAction: 'manipulation' }}
          className={cn(
            'px-3 py-2.5 rounded-md text-sm font-medium transition-colors shrink-0 min-h-[44px]',
            selectedSectionId === 'all'
              ? 'bg-foreground text-background'
              : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
          )}
        >
          All sections
          {totalDiffCount > 0 && (
            <span className={cn(
              'ml-1.5 text-xs font-normal tabular-nums',
              selectedSectionId === 'all' ? 'text-background/60' : 'text-muted-foreground'
            )}>
              {totalDiffCount}
            </span>
          )}
        </button>

        <span className="w-px h-4 bg-border shrink-0" />

        {/* Per-section tabs */}
        {sections.map((section) => {
          const sectionChanges = diffCountBySection.get(section.id) ?? 0
          const color = getSectionColor(section.sortOrder)
          const isActive = selectedSectionId === section.id

          return (
            <button
              key={section.id}
              type="button"
              onClick={() => onSelect(section.id)}
              style={isActive ? { backgroundColor: color, touchAction: 'manipulation' } : { touchAction: 'manipulation' }}
              className={cn(
                'flex items-center gap-2 px-3 py-2.5 rounded-md text-sm font-medium transition-colors shrink-0 min-h-[44px]',
                isActive
                  ? 'text-white'
                  : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
              )}
            >
              <span
                className="w-2 h-2 rounded-full shrink-0"
                style={{ backgroundColor: isActive ? 'rgba(255,255,255,0.75)' : color }}
              />
              <span className="max-w-[160px] truncate">
                {section.name || section.cssSelector.slice(0, 24)}
              </span>
              {sectionChanges > 0 && (
                <span className={cn(
                  'text-xs font-normal tabular-nums shrink-0',
                  isActive ? 'text-white/70' : 'text-muted-foreground'
                )}>
                  {sectionChanges}
                </span>
              )}
            </button>
          )
        })}
      </div>
    </div>
  )
}
