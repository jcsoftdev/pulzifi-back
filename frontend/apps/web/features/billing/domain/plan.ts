/**
 * Billing Feature — Plan Domain Types
 * Plan entity as presented in the UI plan selector
 */

/**
 * A single feature row on a plan card.
 *   - Quantitative features expose `value` (e.g. "10", "Unlimited", "90 days")
 *     and are rendered "Label: value" with a primary ✓ bullet.
 *   - Boolean features expose `included` only and render with ✓ or ✗ bullets.
 *
 * A feature may have BOTH `value` and `included=false` when the tier omits
 * the capability entirely (e.g. AI insights on Starter) — the renderer
 * picks the right visual based on `included`.
 */
export interface PlanFeature {
  label: string
  value?: string
  included: boolean
}

export interface Plan {
  id: string
  name: string
  code: string
  description: string
  /** Monthly price in USD cents */
  priceMonthly: number
  /** Yearly price in USD cents */
  priceYearly: number
  features: PlanFeature[]
  isPopular?: boolean
}

/**
 * Format a price (in USD cents) to a human-readable string.
 * e.g. 2900 → "$29"
 */
export function formatPrice(cents: number): string {
  return `$${Math.round(cents / 100)}`
}
