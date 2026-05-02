export type {
  Integration,
  Target,
  Destination,
  Delivery,
  IntegrationStatus,
  DeliveryStatus,
  ScopeType,
  CreateDestinationInput,
  UpdateDestinationInput,
} from '@workspace/services'

export interface ProviderConfig {
  key: string
  label: string
  description: string
  enabled: boolean // false = coming soon
}

export const PROVIDERS: ProviderConfig[] = [
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
    key: 'teams',
    label: 'Microsoft Teams',
    description: 'Coming soon.',
    enabled: false,
  },
  {
    key: 'discord',
    label: 'Discord',
    description: 'Coming soon.',
    enabled: false,
  },
  {
    key: 'webhook',
    label: 'Webhook',
    description: 'Coming soon.',
    enabled: false,
  },
  {
    key: 'google_sheets',
    label: 'Google Sheets',
    description: 'Coming soon.',
    enabled: false,
  },
]
