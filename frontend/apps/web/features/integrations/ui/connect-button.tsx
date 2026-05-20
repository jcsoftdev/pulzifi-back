'use client'

import { IntegrationsApi } from '@workspace/services'
import Link from 'next/link'
import { useState } from 'react'
import type { Integration } from '../domain/types'
import { TwilioBYOModal } from './twilio-byo-modal'

const PAID_PLANS = [
  'pro',
  'business',
]
const FREE_PLANS = [
  'starter',
  '',
]

interface ConnectButtonProps {
  providerKey: string
  integration?: Integration
  label: string
  planCode?: string
}

export function ConnectButton({
  providerKey,
  integration,
  label,
  planCode = '',
}: Readonly<ConnectButtonProps>) {
  const [byoOpen, setByoOpen] = useState(false)

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

  // Twilio: tier-aware branching
  if (providerKey === 'twilio') {
    const normalizedPlan = planCode.toLowerCase()

    // Free tier: disabled with upgrade prompt
    if (FREE_PLANS.includes(normalizedPlan)) {
      return (
        <button
          type="button"
          disabled
          title="Upgrade your plan to use SMS"
          className="inline-flex items-center justify-center gap-1.5 text-xs font-medium bg-muted text-muted-foreground px-3 py-1.5 rounded-lg opacity-60 cursor-not-allowed"
        >
          Upgrade to use SMS
        </button>
      )
    }

    // Paid tier (pro, business): go straight to destination form
    if (PAID_PLANS.includes(normalizedPlan)) {
      return (
        <Link
          href={`/settings/integrations/${providerKey}`}
          className="inline-flex items-center justify-center gap-1.5 text-xs font-medium bg-primary hover:bg-primary/90 text-primary-foreground px-3 py-1.5 rounded-lg transition-colors"
        >
          Connect {label}
        </Link>
      )
    }

    // Enterprise (or any other plan): BYO modal
    return (
      <>
        <button
          type="button"
          onClick={() => setByoOpen(true)}
          className="inline-flex items-center justify-center gap-1.5 text-xs font-medium bg-primary hover:bg-primary/90 text-primary-foreground px-3 py-1.5 rounded-lg transition-colors"
        >
          Connect {label}
        </button>
        <TwilioBYOModal
          open={byoOpen}
          onClose={() => setByoOpen(false)}
          onConnected={() => {
            // Reload the page so integration list refreshes
            window.location.reload()
          }}
        />
      </>
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
