/**
 * Billing Feature — Plan Domain Types
 * Plan entity as presented in the UI plan selector
 */

export interface Plan {
  id: string
  name: string
  code: string
  description: string
  /** Monthly price in USD cents */
  priceMonthly: number
  /** Yearly price in USD cents */
  priceYearly: number
  features: string[]
  isPopular?: boolean
}

/**
 * Format a price (in USD cents) to a human-readable string.
 * e.g. 2900 → "$29"
 */
export function formatPrice(cents: number): string {
  return `$${Math.round(cents / 100)}`
}
