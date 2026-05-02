'use client'

import { DeliveriesApi } from '@workspace/services'
import type { Delivery, DeliveryStatus } from '../domain/types'
import { useEffect, useState, useRef } from 'react'
import { notification } from '@/lib/notification'

interface DeliveryLogTableProps {
  destinationId: string
}

const STATUS_FILTERS: { value: DeliveryStatus | 'all'; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'delivered', label: 'Delivered' },
  { value: 'pending', label: 'Pending' },
  { value: 'in_flight', label: 'In flight' },
  { value: 'failed', label: 'Failed' },
  { value: 'dead', label: 'Dead' },
]

const STATUS_COLORS: Record<DeliveryStatus, string> = {
  pending: 'bg-yellow-500/15 text-yellow-700 dark:text-yellow-400',
  in_flight: 'bg-blue-500/15 text-blue-700 dark:text-blue-400',
  delivered: 'bg-green-500/15 text-green-700 dark:text-green-400',
  failed: 'bg-red-500/15 text-red-700 dark:text-red-400',
  dead: 'bg-gray-500/15 text-gray-700 dark:text-gray-400',
}

const PAGE_SIZE = 20

function formatDate(iso: string | null | undefined): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function truncate(str: string | null | undefined, len = 60): string {
  if (!str) return '—'
  return str.length > len ? `${str.slice(0, len)}…` : str
}

export function DeliveryLogTable({ destinationId }: Readonly<DeliveryLogTableProps>) {
  const [deliveries, setDeliveries] = useState<Delivery[]>([])
  const [statusFilter, setStatusFilter] = useState<DeliveryStatus | 'all'>('all')
  const [page, setPage] = useState(0)
  const [loading, setLoading] = useState(true)
  const [retrying, setRetrying] = useState<string | null>(null)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const fetchDeliveries = async (silent = false) => {
    if (!silent) setLoading(true)
    try {
      const { data } = await DeliveriesApi.list({
        destination_id: destinationId,
        ...(statusFilter !== 'all' ? { status: statusFilter } : {}),
        limit: PAGE_SIZE,
        offset: page * PAGE_SIZE,
      })
      setDeliveries(data)
    } catch {
      if (!silent) {
        notification.error({ title: 'Failed to load delivery log' })
      }
    } finally {
      if (!silent) setLoading(false)
    }
  }

  // Initial + on filter/page change
  useEffect(() => {
    fetchDeliveries()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [destinationId, statusFilter, page])

  // 10s polling
  useEffect(() => {
    intervalRef.current = setInterval(() => {
      fetchDeliveries(true)
    }, 10_000)
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [destinationId, statusFilter, page])

  const handleRetry = async (id: string) => {
    setRetrying(id)
    try {
      await DeliveriesApi.retry(id)
      notification.success({ title: 'Retry queued' })
      await fetchDeliveries()
    } catch (err) {
      notification.error({
        title: 'Retry failed',
        description: err instanceof Error ? err.message : 'Please try again.',
      })
    } finally {
      setRetrying(null)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-foreground">Delivery log</h3>
        <span className="text-xs text-muted-foreground">Refreshes every 10s</span>
      </div>

      {/* Status filter pills */}
      <div className="flex flex-wrap gap-2">
        {STATUS_FILTERS.map((f) => (
          <button
            key={f.value}
            type="button"
            onClick={() => {
              setStatusFilter(f.value)
              setPage(0)
            }}
            className={`text-xs px-3 py-1 rounded-full border transition-colors ${
              statusFilter === f.value
                ? 'border-primary bg-primary/10 text-primary font-medium'
                : 'border-border text-muted-foreground hover:border-primary/40'
            }`}
          >
            {f.label}
          </button>
        ))}
      </div>

      {/* Table */}
      {loading ? (
        <div className="space-y-2">
          {[1, 2, 3].map((n) => (
            <div key={n} className="h-12 bg-muted/40 rounded-lg animate-pulse" />
          ))}
        </div>
      ) : deliveries.length === 0 ? (
        <p className="text-sm text-muted-foreground py-6 text-center">No deliveries found.</p>
      ) : (
        <div className="overflow-x-auto rounded-xl border border-border">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border bg-muted/30">
                <th className="text-left px-4 py-2.5 text-xs font-medium text-muted-foreground whitespace-nowrap">
                  Date
                </th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-muted-foreground whitespace-nowrap">
                  Event
                </th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-muted-foreground whitespace-nowrap">
                  Status
                </th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-muted-foreground whitespace-nowrap">
                  Attempts
                </th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-muted-foreground">
                  Last error
                </th>
                <th className="px-4 py-2.5" />
              </tr>
            </thead>
            <tbody>
              {deliveries.map((d) => (
                <tr key={d.id} className="border-b border-border/60 hover:bg-muted/20 transition-colors last:border-b-0">
                  <td className="px-4 py-3 text-xs text-muted-foreground whitespace-nowrap">
                    {formatDate(d.createdAt)}
                  </td>
                  <td className="px-4 py-3 text-xs font-mono text-foreground whitespace-nowrap">
                    {d.eventType}
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex items-center text-[11px] font-medium px-2 py-0.5 rounded-full ${STATUS_COLORS[d.status]}`}
                    >
                      {d.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-xs text-muted-foreground text-center">
                    {d.attempts}
                  </td>
                  <td className="px-4 py-3 text-xs text-muted-foreground max-w-xs truncate">
                    {truncate(d.errorMessage)}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {(d.status === 'failed' || d.status === 'dead') && (
                      <button
                        type="button"
                        onClick={() => handleRetry(d.id)}
                        disabled={retrying === d.id}
                        className="text-xs bg-muted hover:bg-muted/70 text-foreground px-2.5 py-1 rounded-md transition-colors disabled:opacity-50"
                      >
                        {retrying === d.id ? '...' : 'Retry'}
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Pagination */}
      <div className="flex items-center justify-between">
        <button
          type="button"
          onClick={() => setPage((p) => Math.max(0, p - 1))}
          disabled={page === 0}
          className="text-xs bg-muted hover:bg-muted/70 text-foreground px-3 py-1.5 rounded-md transition-colors disabled:opacity-40"
        >
          Previous
        </button>
        <span className="text-xs text-muted-foreground">Page {page + 1}</span>
        <button
          type="button"
          onClick={() => setPage((p) => p + 1)}
          disabled={deliveries.length < PAGE_SIZE}
          className="text-xs bg-muted hover:bg-muted/70 text-foreground px-3 py-1.5 rounded-md transition-colors disabled:opacity-40"
        >
          Next
        </button>
      </div>
    </div>
  )
}
