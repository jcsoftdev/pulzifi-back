'use client'

import type { Integration, ProviderState, User } from '@workspace/services'
import { PROVIDER_PRESENTATION } from '../domain/provider-presentation'
import { ProviderCard } from './provider-card'

interface IntegrationsPanelProps {
  integrations: Integration[]
  providers: ProviderState[]
  me?: User | null
}

export function IntegrationsPanel({ integrations, providers, me }: Readonly<IntegrationsPanelProps>) {
  const getIntegration = (key: string): Integration | undefined =>
    integrations.find((i) => i.serviceType === key && i.status === 'active')

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {providers.map((p) => {
        const presentation = PROVIDER_PRESENTATION[p.key] ?? { label: p.key, description: '' }
        return (
          <ProviderCard
            key={p.key}
            providerKey={p.key}
            label={presentation.label}
            description={presentation.description}
            enabled={p.enabled}
            authType={p.authType}
            soon={!p.enabled}
            integration={getIntegration(p.key)}
            planCode={me?.organization?.planCode}
          />
        )
      })}
    </div>
  )
}
