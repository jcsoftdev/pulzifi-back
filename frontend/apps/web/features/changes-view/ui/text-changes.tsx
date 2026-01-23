'use client'

import { cn } from '@workspace/ui/lib/utils'

interface TextChange {
  type: 'added' | 'removed' | 'unchanged'
  text: string
}

export interface TextChangesProps {
  changes?: TextChange[]
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
      <div className="p-6 space-y-4 font-mono text-sm overflow-x-auto">
        {changes.map((change, index) => (
          <div
            key={index}
            className={cn(
              'p-3 rounded-md w-fit whitespace-pre-wrap',
              change.type === 'added' && 'bg-green-500/10 text-green-700 dark:text-green-400',
              change.type === 'removed' && 'bg-red-500/10 text-red-700 dark:text-red-400 line-through decoration-red-500',
              change.type === 'unchanged' && 'text-muted-foreground'
            )}
          >
            {change.text}
          </div>
        ))}
      </div>
    </div>
  )
}
