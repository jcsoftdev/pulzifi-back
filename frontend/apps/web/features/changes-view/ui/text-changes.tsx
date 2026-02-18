'use client'

import { cn } from '@workspace/ui/lib/utils'
import type { DiffRow } from '../utils/simple-diff'

export interface TextChangesProps {
  changes?: DiffRow[]
}

export function TextChanges({ changes = [] }: Readonly<TextChangesProps>) {
  if (!changes || changes.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 bg-muted/20 rounded-lg border border-border">
        <p className="text-muted-foreground">No text changes detected</p>
      </div>
    )
  }

  return (
    <div className="bg-card border border-border rounded-lg overflow-hidden">
      <div className="px-6 py-4 border-b border-border bg-muted/30">
        <h3 className="text-base font-semibold text-foreground">Text Changes</h3>
      </div>

      <div className="p-4 space-y-2 max-h-[480px] overflow-y-auto">
        {changes.map((changeRow, rowIndex) => (
          <div
            key={rowIndex}
            className="px-4 py-3 rounded-lg bg-background border border-border/40"
          >
            <p className="text-sm leading-relaxed">
              {changeRow.segments.map((segment, segmentIndex) => (
                <span
                  key={segmentIndex}
                  className={cn(
                    segment.type === 'removed' && 'line-through text-foreground/50',
                    segment.type === 'added' && 'text-green-600 dark:text-green-400',
                    segment.type === 'unchanged' && 'text-foreground',
                  )}
                >
                  {segmentIndex > 0 ? ' ' : ''}
                  {segment.text}
                </span>
              ))}
            </p>
          </div>
        ))}
      </div>
    </div>
  )
}
