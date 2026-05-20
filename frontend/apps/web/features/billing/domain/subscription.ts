/**
 * Billing Feature — Domain Types
 * Subscription-related entities used in the UI layer
 */

export type BillingStatus =
  | 'active'
  | 'past_due'
  | 'canceled'
  | 'incomplete'
  | 'trialing'

export type BillingCycle = 'monthly' | 'yearly'

export type PaymentStatus = 'ok' | 'failed'

export interface Subscription {
  orgId: string
  stripeSubscriptionId: string
  stripeCustomerId: string
  planId: string
  billingStatus: BillingStatus
  /** ISO-8601 timestamp or null if not yet assigned */
  currentPeriodEnd: string | null
  paymentStatus: PaymentStatus
  updatedAt: string
}

/**
 * Maps a raw API DTO billing_status string to the domain type.
 * Returns 'active' as a safe default for unknown values.
 */
export function toBillingStatus(raw: string): BillingStatus {
  const valid: BillingStatus[] = ['active', 'past_due', 'canceled', 'incomplete', 'trialing']
  return valid.includes(raw as BillingStatus) ? (raw as BillingStatus) : 'active'
}

/**
 * Human-readable label for each billing status.
 */
export function billingStatusLabel(status: BillingStatus): string {
  const labels: Record<BillingStatus, string> = {
    active: 'Active',
    past_due: 'Past Due',
    canceled: 'Canceled',
    incomplete: 'Incomplete',
    trialing: 'Trial',
  }
  return labels[status]
}
