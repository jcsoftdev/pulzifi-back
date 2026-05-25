'use client'

import { Badge } from '@workspace/ui/components/atoms/badge'
import { Button } from '@workspace/ui/components/atoms/button'
import type { Plan } from '../domain/plan'
import { formatPrice } from '../domain/plan'
import type { BillingCycle } from '../domain/subscription'

export type PlanRelation = 'current' | 'upgrade' | 'downgrade' | 'new'

interface PlanCardProps {
  plan: Plan
  billingCycle: BillingCycle
  /** How this plan relates to the user's current subscription. */
  relation: PlanRelation
  /** Enterprise plans bypass checkout/portal and route to a sales contact instead. */
  isEnterprise?: boolean
  isLoading?: boolean
  onChoose: (planId: string, billingCycle: BillingCycle, relation: PlanRelation) => void
}

function ctaLabel(relation: PlanRelation, isEnterprise: boolean, isLoading: boolean): string {
  if (isLoading) return 'Loading...'
  if (isEnterprise) return 'Contact sales'
  switch (relation) {
    case 'current':
      return 'Current plan'
    case 'upgrade':
      return 'Upgrade'
    case 'downgrade':
      return 'Downgrade'
    case 'new':
      return 'Choose plan'
  }
}

function ctaVariant(relation: PlanRelation): 'default' | 'secondary' | 'outline' {
  if (relation === 'current') return 'secondary'
  if (relation === 'downgrade') return 'outline'
  return 'default'
}

/**
 * Presentational card for a single plan in the plan selector.
 * Displays pricing for the selected billing cycle and a context-aware CTA
 * (Current / Upgrade / Downgrade / Choose / Contact sales).
 */
export function PlanCard({
  plan,
  billingCycle,
  relation,
  isEnterprise = false,
  isLoading = false,
  onChoose,
}: PlanCardProps) {
  const isCurrent = relation === 'current'
  const price = billingCycle === 'monthly' ? plan.priceMonthly : plan.priceYearly
  const period = billingCycle === 'monthly' ? '/mo' : '/yr'
  const label = ctaLabel(relation, isEnterprise, isLoading)

  return (
    <div
      className={[
        'relative rounded-2xl border bg-card p-6 flex flex-col gap-4 transition-shadow',
        isCurrent ? 'border-primary shadow-md' : 'border-border',
      ].join(' ')}
    >
      {/* Top badges (current overrides popular) */}
      <div className="absolute -top-3 left-6 flex gap-2">
        {isCurrent && (
          <Badge variant="default" className="text-xs">
            Current plan
          </Badge>
        )}
        {plan.isPopular && !isCurrent && (
          <Badge variant="default" className="text-xs">
            Most popular
          </Badge>
        )}
      </div>

      {/* Plan name */}
      <div>
        <h4 className="text-base font-semibold text-foreground">{plan.name}</h4>
        {plan.description && (
          <p className="text-sm text-muted-foreground mt-0.5">{plan.description}</p>
        )}
      </div>

      {/* Price — Enterprise shows "Custom" instead of a fixed number */}
      <div className="flex items-baseline gap-1">
        {isEnterprise ? (
          <span className="text-3xl font-bold text-foreground">Custom</span>
        ) : (
          <>
            <span className="text-3xl font-bold text-foreground">{formatPrice(price)}</span>
            <span className="text-sm text-muted-foreground">{period}</span>
          </>
        )}
      </div>

      {/* Features */}
      {plan.features.length > 0 && (
        <ul className="space-y-1.5 text-sm text-muted-foreground flex-1">
          {plan.features.map((feature) => (
            <li key={feature} className="flex items-start gap-2">
              <span className="mt-0.5 text-primary">&#10003;</span>
              <span>{feature}</span>
            </li>
          ))}
        </ul>
      )}

      {/* CTA */}
      <Button
        type="button"
        variant={ctaVariant(relation)}
        className="w-full mt-auto"
        disabled={isCurrent || isLoading}
        onClick={() => onChoose(plan.id, billingCycle, relation)}
      >
        {label}
      </Button>
    </div>
  )
}
