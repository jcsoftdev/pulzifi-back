'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { Badge } from '@workspace/ui/components/atoms/badge'
import type { Plan } from '../domain/plan'
import type { BillingCycle } from '../domain/subscription'
import { formatPrice } from '../domain/plan'

interface PlanCardProps {
  plan: Plan
  billingCycle: BillingCycle
  isCurrentPlan?: boolean
  isLoading?: boolean
  onChoose: (planId: string, billingCycle: BillingCycle) => void
}

/**
 * Presentational card for a single plan in the plan selector.
 * Displays pricing for the selected billing cycle.
 */
export function PlanCard({
  plan,
  billingCycle,
  isCurrentPlan = false,
  isLoading = false,
  onChoose,
}: PlanCardProps) {
  const price = billingCycle === 'monthly' ? plan.priceMonthly : plan.priceYearly
  const period = billingCycle === 'monthly' ? '/mo' : '/yr'

  return (
    <div
      className={[
        'relative rounded-2xl border bg-card p-6 flex flex-col gap-4 transition-shadow',
        isCurrentPlan ? 'border-primary shadow-md' : 'border-border',
      ].join(' ')}
    >
      {/* Popular badge */}
      {plan.isPopular && (
        <div className="absolute -top-3 left-6">
          <Badge variant="default" className="text-xs">
            Most popular
          </Badge>
        </div>
      )}

      {/* Plan name */}
      <div>
        <h4 className="text-base font-semibold text-foreground">{plan.name}</h4>
        {plan.description && (
          <p className="text-sm text-muted-foreground mt-0.5">{plan.description}</p>
        )}
      </div>

      {/* Price */}
      <div className="flex items-baseline gap-1">
        <span className="text-3xl font-bold text-foreground">{formatPrice(price)}</span>
        <span className="text-sm text-muted-foreground">{period}</span>
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
        variant={isCurrentPlan ? 'secondary' : 'default'}
        className="w-full mt-auto"
        disabled={isCurrentPlan || isLoading}
        onClick={() => onChoose(plan.id, billingCycle)}
      >
        {isCurrentPlan ? 'Current plan' : isLoading ? 'Loading...' : 'Choose plan'}
      </Button>
    </div>
  )
}
