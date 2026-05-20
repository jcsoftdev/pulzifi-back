'use client'

import type { DeliveryDetail, DeliveryStatus } from '@workspace/services'
import { DeliveriesApi } from '@workspace/services'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@workspace/ui/components/atoms/sheet'
import { useEffect, useState } from 'react'
import { notification } from '@/lib/notification'

const STATUS_COLORS: Record<DeliveryStatus, string> = {
  pending: 'bg-yellow-500/15 text-yellow-700 dark:text-yellow-400',
  in_flight: 'bg-blue-500/15 text-blue-700 dark:text-blue-400',
  delivered: 'bg-green-500/15 text-green-700 dark:text-green-400',
  failed: 'bg-red-500/15 text-red-700 dark:text-red-400',
  dead: 'bg-gray-500/15 text-gray-700 dark:text-gray-400',
}

function formatDate(iso: string | null | undefined): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

interface CollapsibleSectionProps {
  title: string
  children: React.ReactNode
  defaultOpen?: boolean
}

function CollapsibleSection({
  title,
  children,
  defaultOpen = false,
}: Readonly<CollapsibleSectionProps>) {
  const [open, setOpen] = useState(defaultOpen)
  return (
    <div className="border border-border rounded-lg overflow-hidden">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className="w-full flex items-center justify-between px-4 py-2.5 bg-muted/30 hover:bg-muted/50 transition-colors text-left"
      >
        <span className="text-xs font-semibold text-foreground">{title}</span>
        <span className="text-muted-foreground text-xs">{open ? '▲' : '▼'}</span>
      </button>
      {open && <div className="px-4 py-3 bg-background">{children}</div>}
    </div>
  )
}

interface DeliveryDetailDrawerProps {
  deliveryId: string | null
  onClose: () => void
  onRetried?: () => void
}

export function DeliveryDetailDrawer({
  deliveryId,
  onClose,
  onRetried,
}: Readonly<DeliveryDetailDrawerProps>) {
  const [detail, setDetail] = useState<DeliveryDetail | null>(null)
  const [loading, setLoading] = useState(false)
  const [retrying, setRetrying] = useState(false)

  useEffect(() => {
    if (!deliveryId) {
      setDetail(null)
      return
    }
    setLoading(true)
    DeliveriesApi.get(deliveryId)
      .then(setDetail)
      .catch(() => {
        notification.error({
          title: 'Failed to load delivery details',
        })
        onClose()
      })
      .finally(() => setLoading(false))
  }, [
    deliveryId,
    onClose,
  ])

  const handleRetry = async () => {
    if (!detail) return
    setRetrying(true)
    try {
      await DeliveriesApi.retry(detail.id)
      notification.success({
        title: 'Retry queued',
      })
      onRetried?.()
      onClose()
    } catch (err) {
      notification.error({
        title: 'Retry failed',
        description: err instanceof Error ? err.message : 'Please try again.',
      })
    } finally {
      setRetrying(false)
    }
  }

  return (
    <Sheet
      open={Boolean(deliveryId)}
      onOpenChange={(isOpen) => {
        if (!isOpen) onClose()
      }}
    >
      <SheetContent className="w-full sm:max-w-lg overflow-y-auto flex flex-col gap-0 p-0">
        <SheetHeader className="px-6 pt-6 pb-4 border-b border-border">
          <SheetTitle className="text-base font-semibold">Delivery detail</SheetTitle>
          {detail && (
            <SheetDescription asChild>
              <div className="flex items-center gap-2 mt-1">
                <code className="text-xs font-mono text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
                  {detail.id.slice(0, 8)}…
                </code>
                <span
                  className={`inline-flex text-[11px] font-medium px-2 py-0.5 rounded-full ${STATUS_COLORS[detail.status]}`}
                >
                  {detail.status}
                </span>
                <span className="inline-flex text-[11px] font-medium px-2 py-0.5 rounded-full bg-muted text-muted-foreground">
                  {detail.eventType}
                </span>
              </div>
            </SheetDescription>
          )}
        </SheetHeader>

        <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
          {loading && (
            <div className="space-y-3">
              {[
                1,
                2,
                3,
              ].map((n) => (
                <div key={n} className="h-10 bg-muted/40 rounded-lg animate-pulse" />
              ))}
            </div>
          )}

          {!loading && detail && (
            <>
              {/* Timestamps */}
              <div className="grid grid-cols-2 gap-3 text-xs">
                <div>
                  <p className="text-muted-foreground">Created</p>
                  <p className="text-foreground font-medium mt-0.5">
                    {formatDate(detail.createdAt)}
                  </p>
                </div>
                <div>
                  <p className="text-muted-foreground">Last attempt</p>
                  <p className="text-foreground font-medium mt-0.5">
                    {formatDate(detail.lastAttemptAt)}
                  </p>
                </div>
                <div>
                  <p className="text-muted-foreground">Next attempt</p>
                  <p className="text-foreground font-medium mt-0.5">
                    {formatDate(detail.nextAttemptAt)}
                  </p>
                </div>
                <div>
                  <p className="text-muted-foreground">Delivered</p>
                  <p className="text-foreground font-medium mt-0.5">
                    {formatDate(detail.deliveredAt)}
                  </p>
                </div>
                <div>
                  <p className="text-muted-foreground">Attempts</p>
                  <p className="text-foreground font-medium mt-0.5">{detail.attempts}</p>
                </div>
                {detail.responseCode != null && (
                  <div>
                    <p className="text-muted-foreground">Response code</p>
                    <p className="text-foreground font-medium mt-0.5">{detail.responseCode}</p>
                  </div>
                )}
              </div>

              {/* Event payload */}
              <CollapsibleSection title="Event payload">
                <pre className="text-xs font-mono text-foreground overflow-x-auto whitespace-pre-wrap break-all">
                  {JSON.stringify(detail.eventPayload, null, 2)}
                </pre>
              </CollapsibleSection>

              {/* Response body */}
              {detail.responseBody != null && (
                <CollapsibleSection title="Response body">
                  <pre className="text-xs font-mono text-foreground overflow-x-auto whitespace-pre-wrap break-all">
                    {detail.responseBody}
                  </pre>
                </CollapsibleSection>
              )}

              {/* Error message */}
              {detail.errorMessage && (
                <div className="rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3">
                  <p className="text-xs font-medium text-destructive mb-1">Error</p>
                  <p className="text-xs text-foreground">{detail.errorMessage}</p>
                </div>
              )}

              {/* Attempt history */}
              {detail.attemptHistory && detail.attemptHistory.length > 0 && (
                <CollapsibleSection
                  title={`Attempt history (${detail.attemptHistory.length})`}
                  defaultOpen
                >
                  <ol className="space-y-3">
                    {detail.attemptHistory.map((attempt, idx) => (
                      // biome-ignore lint/suspicious/noArrayIndexKey: attempt history items have no unique id
                      <li key={idx} className="flex gap-3 text-xs">
                        <div className="flex flex-col items-center">
                          <div className="w-2 h-2 rounded-full bg-muted-foreground mt-0.5 flex-shrink-0" />
                          {idx < (detail.attemptHistory?.length ?? 0) - 1 && (
                            <div className="w-px flex-1 bg-border mt-1" />
                          )}
                        </div>
                        <div className="pb-3">
                          <p className="text-muted-foreground">{formatDate(attempt.at)}</p>
                          {attempt.code != null && (
                            <p className="text-foreground mt-0.5">
                              Code: <span className="font-mono">{attempt.code}</span>
                            </p>
                          )}
                          {attempt.error && (
                            <p className="text-destructive mt-0.5">{attempt.error}</p>
                          )}
                        </div>
                      </li>
                    ))}
                  </ol>
                </CollapsibleSection>
              )}
            </>
          )}
        </div>

        {/* Footer */}
        {!loading && detail && (
          <div className="px-6 py-4 border-t border-border flex items-center justify-between gap-3">
            <button
              type="button"
              onClick={onClose}
              className="text-xs font-medium bg-muted hover:bg-muted/70 text-foreground px-4 py-2 rounded-lg transition-colors"
            >
              Close
            </button>
            {detail.status === 'dead' && (
              <button
                type="button"
                onClick={handleRetry}
                disabled={retrying}
                className="text-xs font-medium bg-primary hover:bg-primary/90 text-primary-foreground px-4 py-2 rounded-lg transition-colors disabled:opacity-50"
              >
                {retrying ? 'Retrying...' : 'Retry'}
              </button>
            )}
          </div>
        )}
      </SheetContent>
    </Sheet>
  )
}
