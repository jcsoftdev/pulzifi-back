export type ServiceType = 'slack' | 'teams' | 'discord' | 'google_sheets'

export interface IntegrationTab {
  id: ServiceType | 'email' | 'text'
  label: string
  premium: boolean
}

export const INTEGRATION_TABS: IntegrationTab[] = [
  { id: 'email', label: 'Email', premium: false },
  { id: 'text', label: 'Text', premium: true },
  { id: 'slack', label: 'Slack', premium: true },
  { id: 'google_sheets', label: 'Google Sheets', premium: true },
  { id: 'teams', label: 'Teams', premium: true },
]

export const WEBHOOK_INTEGRATIONS: {
  serviceType: ServiceType
  label: string
  placeholder: string
}[] = [
  { serviceType: 'slack', label: 'Slack', placeholder: 'Add the SLACK channel URL' },
  { serviceType: 'google_sheets', label: 'Google Sheets', placeholder: 'Connect to Google Sheets' },
  { serviceType: 'teams', label: 'Teams', placeholder: 'Add the Teams channel URL' },
  { serviceType: 'discord', label: 'Discord', placeholder: 'Add the Discord channel URL' },
]
