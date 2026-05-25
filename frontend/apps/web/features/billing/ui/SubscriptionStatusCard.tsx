'use client'

import type { SubscriptionDto } from '@workspace/services'
import { Badge } from '@workspace/ui/components/atoms/badge'
import type { Plan } from '../domain/plan'
import { formatPrice } from '../domain/plan'
import { billingStatusLabel, toBillingStatus } from '../domain/subscription'
import { ManageBillingButton } from './ManageBillingButton'

interface SubscriptionStatusCardProps {
  subscription: SubscriptionDto
  /** Catalog used to resolve display price for the current plan; matched by code. */
  plans: Plan[]
}

function statusVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (status) {
    case 'active':
    case 'trialing':
      return 'default'
    case 'past_due':
    case 'incomplete':
      return 'secondary'
    case 'canceled':
      return 'destructive'
    default:
      return 'outline'
  }
}

/**
 * Card showing the current subscription: plan name, price, status, period end,
 * and payment health. Falls back gracefully when plan_code is unknown.
 */
export function SubscriptionStatusCard({ subscription, plans }: SubscriptionStatusCardProps) {
  const billingStatus = toBillingStatus(subscription.billing_status ?? 'active')
  const label = billingStatusLabel(billingStatus)

  const currentPlan = plans.find((p) => p.code === subscription.plan_code)
  const planName = subscription.plan_name || currentPlan?.name || 'Unknown plan'
  // Pricing is shown only when we can resolve the plan in the local catalog
  // AND it has a non-zero monthly price (Enterprise / Starter free → hide).
  const monthly = currentPlan?.priceMonthly ?? 0
  const showPrice = monthly > 0

  const periodEnd = subscription.current_period_end
    ? new Date(subscription.current_period_end).toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
      })
    : null

  return (
    <div className="rounded-2xl border border-border bg-card p-6 flex flex-col gap-4">
      {/* Header */}
      <div className="flex items-center justify-between gap-4 flex-wrap">
        <div className="flex items-baseline gap-3 flex-wrap">
          <h3 className="text-base font-semibold text-foreground">{planName}</h3>
          {showPrice && (
            <span className="text-sm text-muted-foreground">{formatPrice(monthly)}/mo</span>
          )}
        </div>
        <Badge variant={statusVariant(billingStatus)}>{label}</Badge>
      </div>

      {/* Details */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
        {periodEnd && (
          <div>
            <p className="text-muted-foreground text-xs uppercase tracking-wide mb-0.5">
              Renews / Ends
            </p>
            <p className="text-foreground font-medium">{periodEnd}</p>
          </div>
        )}

        <div>
          <p className="text-muted-foreground text-xs uppercase tracking-wide mb-0.5">Payment</p>
          {subscription.payment_status === 'ok' || subscription.payment_status === null ? (
            <p className="text-foreground font-medium">Up to date</p>
          ) : (
            <p className="text-destructive font-medium">Action required</p>
          )}
        </div>
      </div>

      {/* Credit balance banner — visible only when customer has store credit */}
      {(subscription.credit_balance_cents ?? 0) > 0 && (
        <div className="rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-3 text-sm">
          <p className="font-medium text-emerald-700 dark:text-emerald-400">
            {formatPrice(subscription.credit_balance_cents ?? 0)}{' '}
            {(subscription.credit_balance_currency || 'usd').toUpperCase()} credit available
          </p>
          <p className="text-muted-foreground text-xs mt-0.5">
            Applied automatically to your next invoice.
          </p>
        </div>
      )}

      {/* Actions */}
      <div className="pt-2 border-t border-border">
        <ManageBillingButton />
      </div>
    </div>
  )
}
