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
  soon?: boolean   // show "Coming soon" badge
}

export const BASE_PROVIDERS: Omit<ProviderConfig, 'enabled' | 'soon'>[] = [
  {
    key: 'slack',
    label: 'Slack',
    description: 'Send alerts to a Slack channel.',
  },
  {
    key: 'email',
    label: 'Email',
    description: 'Send alerts to email recipients.',
  },
  {
    key: 'gmail',
    label: 'Gmail',
    description: 'Send alerts from your Gmail account.',
  },
  {
    key: 'outlook',
    label: 'Outlook',
    description: 'Send alerts from your Outlook / Microsoft 365 account.',
  },
  {
    key: 'discord',
    label: 'Discord',
    description: 'Send alerts to a Discord channel.',
  },
  {
    key: 'twilio',
    label: 'SMS (Twilio)',
    description: 'Send SMS notifications via Twilio.',
  },
  {
    key: 'teams',
    label: 'Microsoft Teams',
    description: 'Coming soon.',
  },
  {
    key: 'google_sheets',
    label: 'Google Sheets',
    description: 'Coming soon.',
  },
]

// Static fallback (all disabled for dynamic providers)
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
    enabled: false,
    soon: true,
  },
  {
    key: 'twilio',
    label: 'SMS (Twilio)',
    description: 'Send SMS notifications via Twilio.',
    enabled: false,
    soon: true,
  },
  {
    key: 'teams',
    label: 'Microsoft Teams',
    description: 'Coming soon.',
    enabled: false,
    soon: true,
  },
  {
    key: 'google_sheets',
    label: 'Google Sheets',
    description: 'Coming soon.',
    enabled: false,
    soon: true,
  },
]
