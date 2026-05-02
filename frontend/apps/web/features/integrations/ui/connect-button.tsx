'use client'

import { IntegrationsApi } from '@workspace/services'
import Link from 'next/link'
import type { Integration } from '../domain/types'

interface ConnectButtonProps {
  providerKey: string
  integration?: Integration
  label: string
  planCode?: string
}

export function ConnectButton({ providerKey, integration, label, planCode: _planCode }: Readonly<ConnectButtonProps>) {
  const handleConnect = () => {
    window.location.href = IntegrationsApi.startOAuthURL(providerKey)
  }

  if (integration) {
    return (
      <Link
        href={`/settings/integrations/${providerKey}`}
        className="inline-flex items-center justify-center gap-1.5 text-xs font-medium bg-muted hover:bg-muted/70 text-foreground px-3 py-1.5 rounded-lg transition-colors"
      >
        Manage
      </Link>
    )
  }

  if (providerKey === 'email') {
    return (
      <Link
        href={`/settings/integrations/${providerKey}`}
        className="inline-flex items-center justify-center gap-1.5 text-xs font-medium bg-primary hover:bg-primary/90 text-primary-foreground px-3 py-1.5 rounded-lg transition-colors"
      >
        Configure {label}
      </Link>
    )
  }

  return (
    <button
      type="button"
      onClick={handleConnect}
      className="inline-flex items-center justify-center gap-1.5 text-xs font-medium bg-primary hover:bg-primary/90 text-primary-foreground px-3 py-1.5 rounded-lg transition-colors"
    >
      Connect {label}
    </button>
  )
}
