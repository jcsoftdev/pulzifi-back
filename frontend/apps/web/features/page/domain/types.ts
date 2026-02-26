export interface Page {
  id: string
  workspaceId: string
  name: string
  url: string
  thumbnailUrl?: string
  lastCheckedAt?: string
  lastChangeDetectedAt?: string
  checkCount: number
  tags: string[]
  checkFrequency: string
  detectedChanges: number
  createdBy: string
  createdAt: string
  updatedAt: string
}

export interface CreatePageDto {
  workspaceId: string
  name: string
  url: string
  tags?: string[]
  checkFrequency?: string
  blockAdsCookies?: boolean
  scheduleType?: string
  enabledInsightTypes?: string[]
  enabledAlertConditions?: string[]
  customAlertCondition?: string
}

export interface EditPageDto {
  name: string
  url: string
  tags?: string[]
  checkFrequency?: string
  blockAdsCookies?: boolean
  scheduleType?: string
  enabledInsightTypes?: string[]
  enabledAlertConditions?: string[]
  customAlertCondition?: string
}

export type PageStatus = 'active' | 'inactive'

export const CHECK_FREQUENCY_OPTIONS = [
  { value: 'Off', label: 'Off' },
  { value: '5m', label: 'Every 5 minutes' },
  { value: '10m', label: 'Every 10 minutes' },
  { value: '15m', label: 'Every 15 minutes' },
  { value: '30m', label: 'Every 30 minutes' },
  { value: '1h', label: 'Every hour' },
  { value: '2h', label: 'Every 2 hours' },
  { value: '4h', label: 'Every 4 hours' },
  { value: '6h', label: 'Every 6 hours' },
  { value: '12h', label: 'Every 12 hours' },
  { value: '24h', label: 'Every day' },
  { value: '168h', label: 'Every week' },
] as const

export type CheckFrequencyValue = (typeof CHECK_FREQUENCY_OPTIONS)[number]['value']

export function getFrequencyLabel(value: string): string {
  return CHECK_FREQUENCY_OPTIONS.find((o) => o.value === value)?.label ?? value
}
