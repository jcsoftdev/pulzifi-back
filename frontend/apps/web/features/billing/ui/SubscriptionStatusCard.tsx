'use client'

import { Badge } from '@workspace/ui/components/atoms/badge'
import type { SubscriptionDto } from '@workspace/services'
import { billingStatusLabel, toBillingStatus } from '../domain/subscription'
import { ManageBillingButton } from './ManageBillingButton'

interface SubscriptionStatusCardProps {
  subscription: SubscriptionDto
}

function statusVariant(
  status: string
): 'default' | 'secondary' | 'destructive' | 'outline' {
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
 * Card showing the current subscription status, period end, and payment health.
 */
export function SubscriptionStatusCard({ subscription }: SubscriptionStatusCardProps) {
  const billingStatus = toBillingStatus(subscription.billing_status ?? 'active')
  const label = billingStatusLabel(billingStatus)

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
        <h3 className="text-base font-semibold text-foreground">Current Subscription</h3>
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
          <p className="text-muted-foreground text-xs uppercase tracking-wide mb-0.5">
            Payment
          </p>
          {subscription.payment_status === 'ok' ? (
            <p className="text-foreground font-medium">Up to date</p>
          ) : (
            <p className="text-destructive font-medium">Action required</p>
          )}
        </div>
      </div>

      {/* Actions */}
      <div className="pt-2 border-t border-border">
        <ManageBillingButton />
      </div>
    </div>
  )
}
