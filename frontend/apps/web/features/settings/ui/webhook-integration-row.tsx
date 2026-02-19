'use client'

import type { Integration } from '@workspace/services'
import { useState } from 'react'
import type { ServiceType } from '../domain/types'

interface WebhookIntegrationRowProps {
  label: string
  serviceType: ServiceType
  placeholder: string
  integration: Integration | undefined
  onConnect: (serviceType: ServiceType, url: string) => Promise<void>
  onDisconnect: (id: string) => Promise<void>
}

export function WebhookIntegrationRow({
  label,
  serviceType,
  placeholder,
  integration,
  onConnect,
  onDisconnect,
}: Readonly<WebhookIntegrationRowProps>) {
  const [url, setUrl] = useState(integration?.config?.url ?? '')
  const [saving, setSaving] = useState(false)

  const handleConnect = async () => {
    if (!url.trim()) return
    setSaving(true)
    try {
      await onConnect(serviceType, url.trim())
    } finally {
      setSaving(false)
    }
  }

  const handleDisconnect = async () => {
    if (!integration) return
    setSaving(true)
    try {
      await onDisconnect(integration.id)
      setUrl('')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="border border-border rounded-xl px-4 py-3 flex items-center gap-3">
      <div className="flex-1">
        <p className="text-xs text-muted-foreground mb-1">{label}</p>
        <input
          type="url"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder={placeholder}
          className="w-full bg-transparent text-sm text-foreground placeholder:text-muted-foreground outline-none"
        />
      </div>
      {integration ? (
        <button
          type="button"
          onClick={handleDisconnect}
          disabled={saving}
          className="flex-shrink-0 text-xs bg-destructive/10 text-destructive hover:bg-destructive/20 px-3 py-1.5 rounded-md transition-colors disabled:opacity-50"
        >
          {saving ? '...' : 'Disconnect'}
        </button>
      ) : (
        <button
          type="button"
          onClick={handleConnect}
          disabled={saving || !url.trim()}
          className="flex-shrink-0 text-xs bg-muted hover:bg-muted/80 text-foreground px-3 py-1.5 rounded-md transition-colors disabled:opacity-50"
        >
          {saving ? '...' : 'Connect'}
        </button>
      )}
    </div>
  )
}
