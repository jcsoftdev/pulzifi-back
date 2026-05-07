'use client'

import type { User } from '@workspace/services'
import type { Integration, ProviderConfig } from '../domain/types'
import { ProviderCard } from './provider-card'

interface IntegrationsPanelProps {
  integrations: Integration[]
  me?: User | null
}

export function IntegrationsPanel({ integrations, me }: Readonly<IntegrationsPanelProps>) {
  const getIntegration = (key: string): Integration | undefined =>
    integrations.find((i) => i.serviceType === key && i.status === 'active')

  const integrationsFlags = (
    me?.organization?.featureFlags as
      | { integrations?: { discord?: boolean; twilio?: boolean; sheets?: boolean; teams?: boolean } }
      | undefined
  )?.integrations

  const discordEnabled = Boolean(integrationsFlags?.discord)
  const twilioEnabled = Boolean(integrationsFlags?.twilio)
  const sheetsEnabled = Boolean(integrationsFlags?.sheets)
  const teamsEnabled = Boolean(integrationsFlags?.teams)

  const providers: ProviderConfig[] = [
    {
      key: 'slack',
      label: 'Slack',
      description: 'Send alerts to a Slack channel.',
      enabled: true,
    },
    {
      key: 'email',
      label: 'Email',
      description: 'Send alerts to email recipients.',
      enabled: true,
    },
    {
      key: 'gmail',
      label: 'Gmail',
      description: 'Send alerts from your Gmail account.',
      enabled: true,
    },
    {
      key: 'outlook',
      label: 'Outlook',
      description: 'Send alerts from your Outlook / Microsoft 365 account.',
      enabled: true,
    },
    {
      key: 'discord',
      label: 'Discord',
      description: 'Send alerts to a Discord channel.',
      enabled: discordEnabled,
      soon: !discordEnabled,
    },
    {
      key: 'twilio',
      label: 'SMS (Twilio)',
      description: 'Send SMS notifications via Twilio.',
      enabled: twilioEnabled,
      soon: !twilioEnabled,
    },
    {
      key: 'sheets',
      label: 'Google Sheets',
      description: 'Append change data to a Google spreadsheet.',
      enabled: sheetsEnabled,
      soon: !sheetsEnabled,
    },
    {
      key: 'teams',
      label: 'Microsoft Teams',
      description: 'Post alerts to a Microsoft Teams channel.',
      enabled: teamsEnabled,
      soon: !teamsEnabled,
    },
  ]

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {providers.map((provider) => (
        <ProviderCard
          key={provider.key}
          providerKey={provider.key}
          label={provider.label}
          description={provider.description}
          enabled={provider.enabled}
          soon={provider.soon}
          integration={getIntegration(provider.key)}
          planCode={me?.organization?.planCode}
        />
      ))}
    </div>
  )
}
