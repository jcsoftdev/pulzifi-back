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
      <div className="p-6 space-y-3 font-mono text-sm max-h-[360px] overflow-y-auto overflow-x-auto">
        {changes.map((changeRow, index) => (
          <div
            key={index}
            className="p-3 rounded-md bg-muted/30 whitespace-pre-wrap"
          >
            {changeRow.segments.map((segment, segmentIndex) => (
              <span
                key={`${index}-${segmentIndex}`}
                className={cn(
                  segment.type === 'added' && 'text-green-700 dark:text-green-400',
                  segment.type === 'removed' && 'text-foreground line-through',
                  segment.type === 'unchanged' && 'text-muted-foreground'
                )}
              >
                {segmentIndex > 0 ? ' ' : ''}
                {segment.text}
              </span>
            ))}
          </div>
        ))}
      </div>
    </div>
  )
}
