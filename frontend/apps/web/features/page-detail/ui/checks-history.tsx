'use client'

import { formatRelativeTime, formatDateTime } from '@workspace/ui'
import { useId } from 'react'
import Link from 'next/link'
import type { Check } from '@workspace/services/page-api'

interface ChecksHistoryProps {
  checks: Check[]
  workspaceId: string
  pageId: string
}

export function ChecksHistory({ checks, workspaceId, pageId }: Readonly<ChecksHistoryProps>) {
  const sectionId = useId()

  const isCheckFailed = (check: Check) => check.status === 'error' || check.status === 'failed'

  return (
    <div id={sectionId} className="flex flex-col gap-6 bg-card border border-border rounded-xl p-6 h-full">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-foreground">Checks history</h2>
      </div>

      <div className="flex flex-col gap-4 max-h-[500px] overflow-y-auto pr-2">
        {/* Today Header - Mock for now as backend doesn't group yet */}
        <div className="flex items-center gap-2">
          <h3 className="text-sm font-medium text-muted-foreground">Today</h3>
          <div className="h-px flex-1 bg-border" />
        </div>

        <div className="relative border-l border-border ml-2 space-y-8">
          {checks.map((check) => (
            <div key={check.id} className="relative pl-6 group">
              {/* Dot */}
              <div
                className={`absolute -left-1.5 top-1.5 h-3 w-3 rounded-full border-2 border-background ${
                  check.changeDetected ? 'bg-destructive' : 'bg-green-100'
                }`}
              />

              <Link 
                href={`/workspaces/${workspaceId}/pages/${pageId}/changes?checkId=${check.id}`}
                className="flex flex-col gap-1 hover:bg-muted/50 p-2 -ml-2 rounded-md transition-colors"
              >
                <span className="text-sm font-medium text-muted-foreground">
                  {formatRelativeTime(check.checkedAt)}
                </span>
                <span className="text-sm text-foreground">{formatDateTime(check.checkedAt)}</span>

                {check.changeDetected ? (
                  <div className="mt-2 inline-flex items-center gap-2 rounded-md border border-destructive/20 bg-destructive/10 px-3 py-1">
                    <span className="text-sm text-destructive">Change found - Alert sent</span>
                    <div className="w-2 h-2 rounded-full bg-destructive" />
                  </div>
                ) : (
                  <div className="mt-1 text-sm text-muted-foreground">No change detected</div>
                )}

                {isCheckFailed(check) && (
                  <div className="mt-2 inline-flex items-center gap-2 rounded-md border border-destructive/20 bg-destructive/10 px-3 py-1">
                    <span className="text-sm text-destructive">
                      {check.extractorFailed
                        ? `Extractor failed${check.errorMessage ? `: ${check.errorMessage}` : ''}`
                        : `Check failed${check.errorMessage ? `: ${check.errorMessage}` : ''}`}
                    </span>
                    <div className="w-2 h-2 rounded-full bg-destructive" />
                  </div>
                )}
              </Link>
            </div>
          ))}
          {checks.length === 0 && (
            <div className="pl-6 text-sm text-muted-foreground">No checks recorded yet.</div>
          )}
        </div>
      </div>
    </div>
  )
}
