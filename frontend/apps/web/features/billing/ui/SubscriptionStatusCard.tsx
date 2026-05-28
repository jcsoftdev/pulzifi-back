'use client'

import type { SubscriptionDto } from '@workspace/services'
import { Badge } from '@workspace/ui/components/atoms/badge'
import type { Plan } from '../domain/plan'
import { formatPrice } from '../domain/plan'
import { billingStatusLabel, toBillingStatus } from '../domain/subscription'
import { CancelPlanButton } from './CancelPlanButton'
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

  const periodEndDate = subscription.current_period_end
    ? new Date(subscription.current_period_end)
    : null
  const periodEnd = periodEndDate
    ? periodEndDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    : null

  // Days remaining until the next billing date — used by the free-month banner
  // so the user sees exactly how long their gift covers.
  const daysRemaining = periodEndDate
    ? Math.max(0, Math.ceil((periodEndDate.getTime() - Date.now()) / (1000 * 60 * 60 * 24)))
    : null

  // Quota for the current plan, pulled from the catalog's "Checks / month" row.
  const checksFeature = currentPlan?.features.find((f) => f.label === 'Checks / month')
  const checksQuota = checksFeature?.value

  // Total prepaid value available to the customer: a gift discount on the
  // subscription PLUS any account credit (e.g. from a duplicate/overpayment).
  // Both translate into months already covered against the current plan price.
  const creditCents = subscription.credit_balance_cents ?? 0
  const giftCents = subscription.gift_amount_off_cents ?? 0
  const coveredCents = creditCents + giftCents
  const hasCoverage = coveredCents > 0
  // How many whole months the available value covers at the current price.
  const monthsCovered = monthly > 0 ? Math.floor(coveredCents / monthly) : 0
  const remainderCents = monthly > 0 ? coveredCents % monthly : coveredCents
  const currency = (subscription.credit_balance_currency || 'usd').toUpperCase()

  // Active plan gift: org is using a higher-tier plan free for a cycle, then
  // reverts. Resolve display names from the catalog (fall back to the code).
  const planGiftActive = subscription.plan_gift_active === true
  const planNameFor = (code?: string) =>
    plans.find((p) => p.code === code)?.name || code || 'your plan'
  const giftPlanName = planNameFor(subscription.gift_plan_code)
  const revertPlanName = planNameFor(subscription.gift_revert_plan_code)
  const giftEndsDate = subscription.gift_ends_at ? new Date(subscription.gift_ends_at) : null
  const giftEnds = giftEndsDate
    ? giftEndsDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    : null
  const giftDaysLeft = giftEndsDate
    ? Math.max(0, Math.ceil((giftEndsDate.getTime() - Date.now()) / (1000 * 60 * 60 * 24)))
    : null

  // A cancellable subscription is one backed by a live Stripe subscription.
  const hasStripeSub = Boolean(subscription.stripe_subscription_id)
  const cancelScheduled = subscription.cancel_at_period_end === true
  const cancelAtDate = subscription.cancel_at ? new Date(subscription.cancel_at) : periodEndDate
  const cancelAt = cancelAtDate
    ? cancelAtDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
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

      {/* Plan-gift banner — org is using a higher plan free for a cycle. Make
          it crystal clear what they have, until when, and what happens next. */}
      {planGiftActive && (
        <div className="rounded-lg border border-violet-500/40 bg-violet-500/10 p-4 text-sm">
          <p className="font-semibold text-violet-700 dark:text-violet-400 flex items-center gap-2">
            <span aria-hidden>🎁</span>
            Free month of {giftPlanName}
          </p>
          <p className="text-muted-foreground mt-1">
            You have full {giftPlanName} access for free
            {giftEnds ? ` until ${giftEnds}` : ''}. After that, your plan returns to{' '}
            {revertPlanName} at its normal price — nothing to do.
          </p>
          {giftDaysLeft !== null && (
            <p className="text-xs text-foreground mt-2">
              <span className="text-muted-foreground">Days left on gift: </span>
              <span className="font-semibold">{giftDaysLeft}</span>
            </p>
          )}
        </div>
      )}

      {/* Coverage banner — unifies gifts and prepaid credit (e.g. a duplicate
          payment) into "months covered". Makes it crystal clear how many
          upcoming invoices are already paid for, plus quota + renewal. */}
      {hasCoverage && (
        <div className="rounded-lg border border-emerald-500/40 bg-emerald-500/10 p-4 text-sm">
          {monthsCovered >= 1 ? (
            <>
              <p className="font-semibold text-emerald-700 dark:text-emerald-400 flex items-center gap-2">
                <span aria-hidden>🎁</span>
                {monthsCovered === 1
                  ? '1 month covered'
                  : `${monthsCovered} months covered`}
              </p>
              <p className="text-muted-foreground mt-1">
                Your next {monthsCovered === 1 ? 'invoice is' : `${monthsCovered} invoices are`}{' '}
                on us — {formatPrice(coveredCents)} {currency} applied automatically
                {remainderCents > 0 ? `, ${formatPrice(remainderCents)} ${currency} left over` : ''}.
              </p>
            </>
          ) : (
            <>
              <p className="font-medium text-emerald-700 dark:text-emerald-400">
                {formatPrice(coveredCents)} {currency} credit available
              </p>
              <p className="text-muted-foreground text-xs mt-0.5">
                Applied automatically to your next invoice.
              </p>
            </>
          )}
          <div className="mt-2 flex flex-wrap gap-x-6 gap-y-1 text-xs text-foreground">
            {daysRemaining !== null && (
              <span>
                <span className="text-muted-foreground">Days left this cycle: </span>
                <span className="font-semibold">{daysRemaining}</span>
              </span>
            )}
            {checksQuota && (
              <span>
                <span className="text-muted-foreground">Included: </span>
                <span className="font-semibold">{checksQuota} checks/month</span>
              </span>
            )}
            {periodEnd && (
              <span>
                <span className="text-muted-foreground">Renews: </span>
                <span className="font-semibold">{periodEnd}</span>
              </span>
            )}
          </div>
        </div>
      )}

      {/* Scheduled cancellation notice */}
      {cancelScheduled && (
        <div className="rounded-lg border border-amber-500/40 bg-amber-500/10 p-4 text-sm">
          <p className="font-semibold text-amber-700 dark:text-amber-400">
            Plan cancellation scheduled
          </p>
          <p className="text-muted-foreground mt-1">
            You'll keep full access{cancelAt ? ` until ${cancelAt}` : ' until the end of this period'}
            , then your plan ends and the workspace becomes read-only. You can resume any time before
            then.
          </p>
        </div>
      )}

      {/* Actions */}
      <div className="pt-2 border-t border-border flex flex-wrap items-start gap-3">
        <ManageBillingButton />
        {hasStripeSub && <CancelPlanButton cancelScheduled={cancelScheduled} />}
      </div>
    </div>
  )
}
