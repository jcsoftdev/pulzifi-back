'use client'

import type { Integration } from '../domain/types'
import { PROVIDERS } from '../domain/types'
import { ProviderCard } from './provider-card'

interface IntegrationsPanelProps {
  integrations: Integration[]
}

export function IntegrationsPanel({ integrations }: Readonly<IntegrationsPanelProps>) {
  const getIntegration = (key: string): Integration | undefined =>
    integrations.find((i) => i.serviceType === key && i.status === 'active')

  return (
    <div className="px-4 md:px-8 lg:px-24 py-8 max-w-4xl">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-foreground">Integrations</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Connect third-party services to receive alerts and notifications.
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {PROVIDERS.map((provider) => (
          <ProviderCard
            key={provider.key}
            providerKey={provider.key}
            label={provider.label}
            description={provider.description}
            enabled={provider.enabled}
            integration={getIntegration(provider.key)}
          />
        ))}
      </div>
    </div>
  )
}
