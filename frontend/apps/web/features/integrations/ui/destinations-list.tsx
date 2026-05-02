'use client'

import { DestinationsApi } from '@workspace/services'
import type { Destination } from '../domain/types'
import { useState } from 'react'
import { notification } from '@/lib/notification'
import { DestinationForm } from './destination-form'

interface DestinationsListProps {
  providerKey: string
  integrationId?: string
  scopeType: 'org' | 'workspace' | 'page'
  scopeId: string
  destinations: Destination[]
  onUpdated: (destinations: Destination[]) => void
}

function destinationSummary(d: Destination): string {
  if (d.serviceType === 'slack') {
    const name = ((d.target?.channel_name as string | undefined) ?? (d.target?.channel_id as string | undefined)) ?? '—'
    return `#${name}`
  }
  if (d.serviceType === 'email') {
    const emails = d.target?.emails as string[] | undefined
    if (emails && emails.length > 0) {
      const first = emails[0] ?? ''
      return emails.length === 1 ? first : `${first} +${emails.length - 1} more`
    }
  }
  return d.serviceType
}

export function DestinationsList({
  providerKey,
  integrationId,
  scopeType,
  scopeId,
  destinations,
  onUpdated,
}: Readonly<DestinationsListProps>) {
  const [editingId, setEditingId] = useState<string | null>(null)
  const [deletingId, setDeletingId] = useState<string | null>(null)

  const handleDelete = async (id: string) => {
    setDeletingId(id)
    try {
      await DestinationsApi.delete(id)
      onUpdated(destinations.filter((d) => d.id !== id))
      notification.success({ title: 'Destination removed' })
    } catch (err) {
      notification.error({
        title: 'Failed to remove destination',
        description: err instanceof Error ? err.message : 'Please try again.',
      })
    } finally {
      setDeletingId(null)
    }
  }

  const handleUpdated = (updated: Destination) => {
    onUpdated(destinations.map((d) => (d.id === updated.id ? updated : d)))
    setEditingId(null)
  }

  if (destinations.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">No destinations yet. Add one above.</p>
    )
  }

  return (
    <div className="space-y-3">
      {destinations.map((d) => (
        <div key={d.id} className="rounded-xl border border-border bg-card p-4">
          {editingId === d.id ? (
            <>
              <DestinationForm
                providerKey={providerKey}
                integrationId={integrationId}
                scopeType={scopeType}
                scopeId={scopeId}
                destination={d}
                onSuccess={handleUpdated}
              />
              <button
                type="button"
                onClick={() => setEditingId(null)}
                className="mt-2 text-xs text-muted-foreground hover:underline"
              >
                Cancel
              </button>
            </>
          ) : (
            <div className="flex items-center justify-between gap-3">
              <div className="min-w-0">
                <p className="text-sm font-medium text-foreground truncate">
                  {destinationSummary(d)}
                </p>
                <p className="text-xs text-muted-foreground mt-0.5">
                  Events: {d.events.join(', ')}
                </p>
              </div>
              <div className="flex items-center gap-2 flex-shrink-0">
                <button
                  type="button"
                  onClick={() => setEditingId(d.id)}
                  className="text-xs bg-muted hover:bg-muted/70 text-foreground px-2.5 py-1.5 rounded-md transition-colors"
                >
                  Edit
                </button>
                <button
                  type="button"
                  onClick={() => handleDelete(d.id)}
                  disabled={deletingId === d.id}
                  className="text-xs bg-destructive/10 text-destructive hover:bg-destructive/20 px-2.5 py-1.5 rounded-md transition-colors disabled:opacity-50"
                >
                  {deletingId === d.id ? '...' : 'Remove'}
                </button>
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}
