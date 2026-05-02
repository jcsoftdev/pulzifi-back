'use client'

import { DestinationsApi } from '@workspace/services'
import type { CreateDestinationInput, Destination, ScopeType, Target } from '../domain/types'
import { useCallback, useEffect, useState } from 'react'
import { notification } from '@/lib/notification'

const AVAILABLE_EVENTS = [
  { value: 'change.detected', label: 'Change detected' },
  { value: 'alert.created', label: 'Alert created' },
]

const E164_REGEX = /^\+[1-9]\d{6,14}$/

interface DestinationFormProps {
  providerKey: string
  integrationId?: string
  scopeType: ScopeType
  scopeId: string
  /** Existing destination — when set, we're in edit mode. */
  destination?: Destination
  onSuccess: (destination: Destination) => void
}

export function DestinationForm({
  providerKey,
  integrationId,
  scopeType,
  scopeId,
  destination,
  onSuccess,
}: Readonly<DestinationFormProps>) {
  const isEdit = Boolean(destination)

  // Slack: channel selected from dropdown
  const [targets, setTargets] = useState<Target[]>([])
  const [selectedTarget, setSelectedTarget] = useState<string>(
    destination?.target?.channel_id as string ?? ''
  )
  const [loadingTargets, setLoadingTargets] = useState(false)

  // Email: comma-separated list
  const [emailInput, setEmailInput] = useState<string>(
    destination?.target?.emails
      ? (destination.target.emails as string[]).join(', ')
      : ''
  )

  // Twilio: comma-separated phone numbers
  const [phoneInput, setPhoneInput] = useState<string>(
    destination?.target?.phone_numbers
      ? (destination.target.phone_numbers as string[]).join(', ')
      : ''
  )
  const [phoneErrors, setPhoneErrors] = useState<string[]>([])

  // Events checkboxes
  const [selectedEvents, setSelectedEvents] = useState<string[]>(
    destination?.events ?? ['change.detected', 'alert.created']
  )

  const [saving, setSaving] = useState(false)

  const fetchTargets = useCallback(async () => {
    if (providerKey !== 'slack' || !integrationId) return
    setLoadingTargets(true)
    try {
      const list = await import('@workspace/services').then((m) =>
        m.IntegrationsApi.listTargets(integrationId)
      )
      setTargets(list)
    } catch {
      // Non-fatal: if targets fail, user can't select a channel but form still shows
    } finally {
      setLoadingTargets(false)
    }
  }, [providerKey, integrationId])

  useEffect(() => {
    fetchTargets()
  }, [fetchTargets])

  const toggleEvent = (value: string) => {
    setSelectedEvents((prev) =>
      prev.includes(value) ? prev.filter((e) => e !== value) : [...prev, value]
    )
  }

  const parsePhones = (): string[] =>
    phoneInput
      .split(',')
      .map((p) => p.trim())
      .filter(Boolean)

  const validatePhones = (phones: string[]): string[] =>
    phones.filter((p) => !E164_REGEX.test(p))

  const buildTarget = (): Record<string, unknown> => {
    if (providerKey === 'slack') {
      const found = targets.find((t) => t.id === selectedTarget)
      return { channel_id: selectedTarget, channel_name: found?.name ?? '' }
    }
    if (providerKey === 'twilio') {
      const phones = parsePhones()
      return { phone_numbers: phones }
    }
    // email
    const emails = emailInput
      .split(',')
      .map((e) => e.trim())
      .filter(Boolean)
    return { emails }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (selectedEvents.length === 0) {
      notification.error({ title: 'Select at least one event' })
      return
    }

    const target = buildTarget()

    if (providerKey === 'slack' && !selectedTarget) {
      notification.error({ title: 'Select a Slack channel' })
      return
    }
    if (providerKey === 'email') {
      const emails = (target.emails as string[]) ?? []
      if (emails.length === 0) {
        notification.error({ title: 'Enter at least one email address' })
        return
      }
    }
    if (providerKey === 'twilio') {
      const phones = parsePhones()
      if (phones.length === 0) {
        notification.error({ title: 'Enter at least one phone number' })
        return
      }
      const invalid = validatePhones(phones)
      if (invalid.length > 0) {
        setPhoneErrors(invalid)
        notification.error({
          title: 'Invalid phone numbers',
          description: `These numbers are not in E.164 format: ${invalid.join(', ')}`,
        })
        return
      }
      setPhoneErrors([])
    }

    setSaving(true)
    try {
      let result: Destination
      if (isEdit && destination) {
        result = await DestinationsApi.update(destination.id, {
          target,
          events: selectedEvents,
        })
        notification.success({ title: 'Destination updated' })
      } else {
        const body: CreateDestinationInput = {
          service_type: providerKey,
          scope_type: scopeType,
          scope_id: scopeId,
          target,
          events: selectedEvents,
          ...(integrationId ? { integration_id: integrationId } : {}),
        }
        result = await DestinationsApi.create(body)
        notification.success({ title: 'Destination created' })
      }
      onSuccess(result)
    } catch (err) {
      notification.error({
        title: isEdit ? 'Failed to update destination' : 'Failed to create destination',
        description: err instanceof Error ? err.message : 'Please try again.',
      })
    } finally {
      setSaving(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {/* Slack: channel selector */}
      {providerKey === 'slack' && (
        <div>
          <label className="text-xs font-medium text-muted-foreground block mb-1.5">
            Slack channel
          </label>
          {loadingTargets ? (
            <div className="h-9 bg-muted rounded-lg animate-pulse" />
          ) : (
            <select
              value={selectedTarget}
              onChange={(e) => setSelectedTarget(e.target.value)}
              className="w-full h-9 rounded-lg border border-border bg-background px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
            >
              <option value="">Select a channel...</option>
              {targets.map((t) => (
                <option key={t.id} value={t.id}>
                  #{t.name}
                </option>
              ))}
            </select>
          )}
        </div>
      )}

      {/* Email: recipients */}
      {providerKey === 'email' && (
        <div>
          <label className="text-xs font-medium text-muted-foreground block mb-1.5">
            Email recipients
          </label>
          <textarea
            value={emailInput}
            onChange={(e) => setEmailInput(e.target.value)}
            placeholder="alice@example.com, bob@example.com"
            rows={2}
            className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 resize-none"
          />
          <p className="text-xs text-muted-foreground mt-1">Separate multiple addresses with commas.</p>
        </div>
      )}

      {/* Twilio: phone numbers */}
      {providerKey === 'twilio' && (
        <div>
          <label className="text-xs font-medium text-muted-foreground block mb-1.5">
            Phone numbers (E.164)
          </label>
          <textarea
            value={phoneInput}
            onChange={(e) => {
              setPhoneInput(e.target.value)
              setPhoneErrors([])
            }}
            placeholder="+15551234567, +447700900123"
            rows={2}
            className={`w-full rounded-lg border bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 resize-none ${
              phoneErrors.length > 0
                ? 'border-destructive focus:ring-destructive/40'
                : 'border-border focus:ring-primary/40'
            }`}
          />
          <p className="text-xs text-muted-foreground mt-1">
            Separate multiple numbers with commas. Must be in E.164 format (e.g. +15551234567).
          </p>
          {phoneErrors.length > 0 && (
            <ul className="mt-1 space-y-0.5">
              {phoneErrors.map((p) => (
                <li key={p} className="text-xs text-destructive">
                  {p} — invalid format
                </li>
              ))}
            </ul>
          )}
        </div>
      )}

      {/* Events */}
      <div>
        <label className="text-xs font-medium text-muted-foreground block mb-1.5">
          Notify on
        </label>
        <div className="flex flex-wrap gap-3">
          {AVAILABLE_EVENTS.map((ev) => (
            <label key={ev.value} className="flex items-center gap-2 cursor-pointer select-none">
              <input
                type="checkbox"
                checked={selectedEvents.includes(ev.value)}
                onChange={() => toggleEvent(ev.value)}
                className="rounded border-border text-primary focus:ring-primary/40"
              />
              <span className="text-sm text-foreground">{ev.label}</span>
            </label>
          ))}
        </div>
      </div>

      <button
        type="submit"
        disabled={saving}
        className="inline-flex items-center gap-1.5 text-sm font-medium bg-primary hover:bg-primary/90 text-primary-foreground px-4 py-2 rounded-lg transition-colors disabled:opacity-50"
      >
        {saving ? 'Saving...' : isEdit ? 'Update destination' : 'Add destination'}
      </button>
    </form>
  )
}
