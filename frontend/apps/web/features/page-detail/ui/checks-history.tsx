'use client'

import {
  type Check,
  type CheckBackendDto,
  mapBackendCheck,
  PageApi,
} from '@workspace/services/page-api'
import { refreshAndRetry } from '@workspace/shared-http'
import { formatDateTime, formatRelativeTime } from '@workspace/ui'
import { Loader2 } from 'lucide-react'
import Link from 'next/link'
import { useCallback, useEffect, useId, useRef, useState } from 'react'

const MAX_SSE_RETRIES = 5

function isCheckInProgress(check: Check) {
  return check.status === 'pending' || check.status === 'running'
}

interface ChecksHistoryProps {
  checks: Check[]
  workspaceId: string
  pageId: string
  quotaExceeded?: boolean
  refillDate?: string
}

export function ChecksHistory({
  checks: initialChecks,
  workspaceId,
  pageId,
  quotaExceeded,
  refillDate,
}: Readonly<ChecksHistoryProps>) {
  const sectionId = useId()
  const [checks, setChecks] = useState(initialChecks)
  const [hasFreshData, setHasFreshData] = useState(false)
  const eventSourceRef = useRef<EventSource | null>(null)
  const sseRetryCount = useRef(0)

  // Sync server-rendered data when it changes (e.g. after router.refresh() or navigation).
  const initialSyncKey = `${initialChecks.length}:${initialChecks[0]?.id}`
  const [prevSyncKey, setPrevSyncKey] = useState(initialSyncKey)
  if (initialSyncKey !== prevSyncKey) {
    setPrevSyncKey(initialSyncKey)
    setChecks(initialChecks)
    setHasFreshData(true)
  }

  const fetchChecks = useCallback(async () => {
    try {
      const freshChecks = await PageApi.listChecks(pageId)
      setChecks((prev) => {
        // Preserve in-progress checks delivered via SSE that aren't in the API response yet
        // (the check record may not be committed to DB when this fetch runs).
        const freshIds = new Set(freshChecks.map((c) => c.id))
        const sseOnlyInProgress = prev.filter(
          (c) => isCheckInProgress(c) && !freshIds.has(c.id),
        )
        return sseOnlyInProgress.length > 0
          ? [...sseOnlyInProgress, ...freshChecks]
          : freshChecks
      })
      setHasFreshData(true)
    } catch (error) {
      console.error('Failed to fetch checks', error)
      setHasFreshData(true)
    }
  }, [pageId])

  // Connect (or reconnect) the EventSource for real-time check updates.
  const connectSSE = useCallback(() => {
    eventSourceRef.current?.close()

    const es = new EventSource(
      `/api/v1/monitoring/checks/page/${pageId}/stream`,
    )
    eventSourceRef.current = es

    es.addEventListener('check:updated', (event) => {
      try {
        const dto: CheckBackendDto = JSON.parse(event.data)
        // Skip section checks — only parent/full-page checks belong in the history list.
        if (dto.section_id) return
        const updated = mapBackendCheck(dto)
        setChecks((prev) => {
          const idx = prev.findIndex((c) => c.id === updated.id)
          const next = idx >= 0
            ? prev.map((c, i) => (i === idx ? updated : c))
            : [updated, ...prev]
          // Notify siblings after React finishes this render batch.
          const hasActive = next.some(isCheckInProgress)
          queueMicrotask(() => {
            window.dispatchEvent(new CustomEvent('checks:active', { detail: hasActive }))
          })
          return next
        })
        setHasFreshData(true)
      } catch {
        // Ignore malformed events
      }
    })

    es.onopen = () => {
      sseRetryCount.current = 0
    }

    es.onerror = async () => {
      if (es.readyState === EventSource.CLOSED) {
        // Connection permanently closed (likely 401 from expired JWT).
        es.close()
        eventSourceRef.current = null

        if (sseRetryCount.current < MAX_SSE_RETRIES) {
          sseRetryCount.current++
          // Reuse shared refresh utility (deduplicates with Axios interceptor,
          // redirects to login on failure).
          const refreshed = await refreshAndRetry()
          if (refreshed) {
            connectSSE()
          }
        }
      }
      // CONNECTING state = browser auto-retry, do nothing.
    }

    return es
  }, [pageId])

  // SSE lifecycle
  useEffect(() => {
    connectSSE()
    return () => {
      eventSourceRef.current?.close()
      eventSourceRef.current = null
    }
  }, [connectSSE])

  // Fetch fresh data on mount and when page becomes visible again.
  useEffect(() => {
    setHasFreshData(false)
    fetchChecks()

    const onVisible = () => {
      if (document.visibilityState === 'visible') {
        fetchChecks()
        // Reconnect SSE in case it died while tab was hidden.
        if (!eventSourceRef.current || eventSourceRef.current.readyState === EventSource.CLOSED) {
          sseRetryCount.current = 0
          connectSSE()
        }
      }
    }
    document.addEventListener('visibilitychange', onVisible)
    return () => document.removeEventListener('visibilitychange', onVisible)
  }, [fetchChecks, connectSSE])

  // Listen for "checks:refresh" (fired by Run Now) — single fetch as safety net.
  useEffect(() => {
    const handler = () => fetchChecks()
    window.addEventListener('checks:refresh', handler)
    return () => window.removeEventListener('checks:refresh', handler)
  }, [fetchChecks])

  const isCheckFailed = (check: Check) =>
    check.status === 'error' || check.status === 'failed'

  return (
    <div
      id={sectionId}
      className="flex flex-col gap-6 bg-card border border-border rounded-xl p-6 h-full"
    >
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-foreground">
          Checks history
        </h2>
      </div>

      <div className="flex flex-col gap-4 max-h-[500px] overflow-y-auto pr-2">
        {quotaExceeded && (
          <div className="rounded-md border border-yellow-500/30 bg-yellow-500/10 px-4 py-3 text-sm text-yellow-700 dark:text-yellow-400">
            Monthly check quota reached. New checks are paused until {refillDate ?? 'next billing cycle'}.
          </div>
        )}

        {/* Today Header - Mock for now as backend doesn't group yet */}
        <div className="flex items-center gap-2">
          <h3 className="text-sm font-medium text-muted-foreground">Today</h3>
          <div className="h-px flex-1 bg-border" />
        </div>

        <div className="relative border-l border-border ml-2 space-y-8">
          {checks.map((check) => {
            // Only trust "in progress" status from fresh API data,
            // not stale server-rendered props after back-navigation
            const inProgress = hasFreshData && isCheckInProgress(check)

            const content = (
              <>
                <span className="text-sm font-medium text-muted-foreground">
                  {formatRelativeTime(check.checkedAt)}
                </span>
                <span className="text-sm text-foreground">
                  {formatDateTime(check.checkedAt)}
                </span>

                {inProgress ? (
                  <div className="mt-2 inline-flex items-center gap-2 rounded-md border border-primary/20 bg-primary/10 px-3 py-1">
                    <Loader2 className="w-3 h-3 animate-spin text-primary" />
                    <span className="text-sm text-primary">
                      Running check...
                    </span>
                  </div>
                ) : check.changeDetected ? (
                  <div className="mt-2 inline-flex items-center gap-2 rounded-md border border-destructive/20 bg-destructive/10 px-3 py-1">
                    <span className="text-sm text-destructive">
                      Change found - Alert sent
                    </span>
                    <div className="w-2 h-2 rounded-full bg-destructive" />
                  </div>
                ) : isCheckFailed(check) ? (
                  <div className="mt-2 inline-flex items-center gap-2 rounded-md border border-destructive/20 bg-destructive/10 px-3 py-1">
                    <span className="text-sm text-destructive">
                      {check.extractorFailed
                        ? `Extractor failed${check.errorMessage ? `: ${check.errorMessage}` : ''}`
                        : `Check failed${check.errorMessage ? `: ${check.errorMessage}` : ''}`}
                    </span>
                    <div className="w-2 h-2 rounded-full bg-destructive" />
                  </div>
                ) : (
                  <div className="mt-1 text-sm text-muted-foreground">
                    No change detected
                  </div>
                )}
              </>
            )

            return (
              <div key={check.id} className="relative pl-6 group">
                {/* Dot — pulsing animation while in progress */}
                <div
                  className={`absolute -left-1.5 top-1.5 h-3 w-3 rounded-full border-2 border-background ${
                    inProgress
                      ? 'bg-primary animate-pulse'
                      : check.changeDetected
                        ? 'bg-destructive'
                        : isCheckFailed(check)
                          ? 'bg-destructive'
                          : 'bg-green-100'
                  }`}
                />

                {inProgress ? (
                  <div className="flex flex-col gap-1 p-2 -ml-2 rounded-md">
                    {content}
                  </div>
                ) : (
                  <Link
                    href={`/workspaces/${workspaceId}/pages/${pageId}/changes?checkId=${check.id}`}
                    className="flex flex-col gap-1 hover:bg-muted/50 p-2 -ml-2 rounded-md transition-colors"
                  >
                    {content}
                  </Link>
                )}
              </div>
            )
          })}
          {checks.length === 0 && (
            <div className="pl-6 text-sm text-muted-foreground">
              No checks recorded yet.
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
