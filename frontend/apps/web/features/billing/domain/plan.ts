/**
 * Billing Feature — Plan Domain Types
 */
import type { PlanDto } from '@workspace/services'

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

/** Format a price (USD cents) to a human-readable string. e.g. 2900 → "$29" */
export function formatPrice(cents: number): string {
  return `$${Math.round(cents / 100)}`
}

// Boolean features per plan code — product decisions, not DB-driven.
const BOOL_FEATURES: ReadonlyArray<{
  label: string
  included: Record<string, boolean>
}> = [
  { label: 'Slack & email alerts', included: { starter: false, pro: true,  enterprise: true  } },
  { label: 'SLA guarantee',        included: { starter: false, pro: false, enterprise: true  } },
  { label: 'Priority support',     included: { starter: false, pro: false, enterprise: true  } },
]

function featuresFor(dto: PlanDto): PlanFeature[] {
  const quant: PlanFeature[] = [
    {
      label: 'Workspaces',
      value: dto.max_workspaces === null ? 'Unlimited' : String(dto.max_workspaces),
      included: true,
    },
    {
      label: 'Pages',
      value: dto.max_pages === null ? 'Custom' : String(dto.max_pages),
      included: true,
    },
    {
      label: 'Checks / month',
      value: dto.checks_allowed_monthly === null
        ? 'Unlimited'
        : dto.checks_allowed_monthly.toLocaleString(),
      included: true,
    },
    {
      label: 'History',
      value: `${dto.storage_period_days} days`,
      included: true,
    },
    {
      label: 'AI insights / month',
      value: dto.ai_insights_allowed_monthly === null ? 'Unlimited' : String(dto.ai_insights_allowed_monthly),
      included: true,
    },
  ]

  const bools: PlanFeature[] = BOOL_FEATURES.map((f) => ({
    label: f.label,
    included: f.included[dto.code] ?? false,
  }))

  return [...quant, ...bools]
}

/** Map a backend PlanDto to the frontend Plan entity. */
export function toPlan(dto: PlanDto): Plan {
  return {
    id:           dto.id,
    name:         dto.name,
    code:         dto.code,
    description:  dto.description,
    priceMonthly: dto.price_monthly_cents,
    priceYearly:  dto.price_yearly_cents,
    features:     featuresFor(dto),
    isPopular:    dto.is_popular,
  }
}
