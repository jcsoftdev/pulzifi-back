'use client'

import { cn } from '@workspace/ui/lib/utils'
import type { WorkspaceStatus } from '../domain/types'

export interface TabsFilterProps {
  selected: WorkspaceStatus
  onChange: (status: WorkspaceStatus) => void
}

const tabs: {
  label: string
  value: WorkspaceStatus
}[] = [
  {
    label: 'Active',
    value: 'Active',
  },
  {
    label: 'Deleted',
    value: 'Deleted',
  },
]

export function TabsFilter({ selected, onChange }: Readonly<TabsFilterProps>) {
  return (
    <div className="flex border-b border-border">
      {tabs.map((tab) => (
        <button
          type="button"
          key={tab.value}
          onClick={() => onChange(tab.value)}
          className={cn(
            'px-4 py-2 text-sm font-semibold transition-colors',
            selected === tab.value
              ? 'text-foreground border-b-2 border-foreground -mb-px'
              : 'text-muted-foreground hover:text-foreground'
          )}
        >
          {tab.label}
        </button>
      ))}
    </div>
  )
}
