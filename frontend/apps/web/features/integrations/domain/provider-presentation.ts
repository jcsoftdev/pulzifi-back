// Presentation-only metadata for providers. The backend owns capability/state
// (enabled, connected, authType); the frontend owns labels/descriptions/icons.
export interface ProviderPresentation {
  label: string
  description: string
}

export const PROVIDER_PRESENTATION: Record<string, ProviderPresentation> = {
  slack: { label: 'Slack', description: 'Send alerts to a Slack channel.' },
  email: { label: 'Email', description: 'Send alerts to email recipients.' },
  gmail: { label: 'Gmail', description: 'Send alerts from your Gmail account.' },
  outlook: {
    label: 'Outlook',
    description: 'Send alerts from your Outlook / Microsoft 365 account.',
  },
  discord: { label: 'Discord', description: 'Send alerts to a Discord channel.' },
  twilio: { label: 'SMS (Twilio)', description: 'Send SMS notifications via Twilio.' },
  sheets: { label: 'Google Sheets', description: 'Append change data to a Google spreadsheet.' },
  teams: { label: 'Microsoft Teams', description: 'Post alerts to a Microsoft Teams channel.' },
}
