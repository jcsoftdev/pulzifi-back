'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { useState } from 'react'
import { useCreateCheckout } from '../application/useCreateCheckout'
import { useCustomerPortal } from '../application/useCustomerPortal'
import { useSubscription } from '../application/useSubscription'
import type { Plan } from '../domain/plan'
import type { BillingCycle } from '../domain/subscription'
import { PlanCard, type PlanRelation } from './PlanCard'
import { SubscriptionStatusCard } from './SubscriptionStatusCard'

/**
 * Static plan catalog.
 * Order matters — index drives upgrade/downgrade detection (`PLANS[i].code` vs
 * `subscription.plan_code`). When fetching plans from an API later, preserve
 * tier order: starter → pro → enterprise.
 */
const PLANS: Plan[] = [
  {
    id: 'starter',
    name: 'Starter',
    code: 'starter',
    description: 'For small teams getting started with monitoring.',
    priceMonthly: 0,
    priceYearly: 0,
    features: ['3 workspaces', '10 pages', '500 checks/month', '7-day history'],
  },
  {
    id: 'pro',
    name: 'Pro',
    code: 'pro',
    description: 'Everything you need to monitor at scale.',
    priceMonthly: 2900,
    priceYearly: 29000,
    features: [
      'Unlimited workspaces',
      '100 pages',
      '10,000 checks/month',
      '90-day history',
      'AI insights',
      'Slack & email alerts',
    ],
    isPopular: true,
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    code: 'enterprise',
    description: 'Custom limits, SLA, and dedicated support.',
    priceMonthly: 0,
    priceYearly: 0,
    features: [
      'Everything in Pro',
      'Custom page limits',
      'Unlimited checks',
      'SLA guarantee',
      'Priority support',
    ],
  },
]

const SALES_EMAIL = 'sales@pulzifi.com'

function relationFor(
  planCode: string,
  currentTierIndex: number,
  thisTierIndex: number,
  currentCycle: BillingCycle | '' | undefined,
  selectedCycle: BillingCycle
): PlanRelation {
  if (currentTierIndex === -1) return 'new'
  if (thisTierIndex === currentTierIndex) {
    // Same tier — if the selected billing cycle differs from the active one,
    // the user is actually changing cycles. Treat that as an upgrade so the
    // CTA routes to the Customer Portal (where Stripe handles the swap).
    if (currentCycle && currentCycle !== selectedCycle) return 'upgrade'
    return 'current'
  }
  if (thisTierIndex > currentTierIndex) return 'upgrade'
  return 'downgrade'
}

/**
 * Main billing tab — shows current subscription + plan selector with monthly/yearly toggle.
 * Handles the BillingEnabled=false case by catching 404s in useSubscription (returns null).
 *
 * CTA routing:
 *   - No active sub → checkout (new Stripe session)
 *   - Active sub, different plan → Stripe Customer Portal (in-place upgrade/downgrade w/ proration)
 *   - Enterprise → mailto sales (no Stripe flow)
 */
export function BillingTab() {
  const { subscription, isLoading: subLoading, error: subError } = useSubscription()
  const { isLoading: checkoutLoading, error: checkoutError, createCheckout } = useCreateCheckout()
  const { isLoading: portalLoading, error: portalError, openPortal } = useCustomerPortal()
  const [billingCycle, setBillingCycle] = useState<BillingCycle>('monthly')

  const currentTierIndex = subscription?.plan_code
    ? PLANS.findIndex((p) => p.code === subscription.plan_code)
    : -1

  const actionLoading = checkoutLoading || portalLoading
  const actionError = checkoutError || portalError

  const handleChoosePlan = (planId: string, cycle: BillingCycle, relation: PlanRelation) => {
    const plan = PLANS.find((p) => p.id === planId)
    if (plan?.code === 'enterprise') {
      window.location.href = `mailto:${SALES_EMAIL}?subject=Enterprise plan inquiry`
      return
    }
    // Portal works whenever the org has a Stripe customer — even if our DB
    // has not yet recorded the subscription_id (e.g. webhook lag/replay).
    // First-time paid pick (no customer yet) must go through Checkout.
    const hasStripeCustomer = Boolean(subscription?.stripe_customer_id)
    if (hasStripeCustomer && (relation === 'upgrade' || relation === 'downgrade')) {
      openPortal()
      return
    }
    createCheckout(planId, cycle)
  }

  return (
    <div className="px-4 md:px-8 lg:px-16 py-8 max-w-5xl">
      {/* Page header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-foreground">Billing</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Manage your subscription and payment method.
        </p>
      </div>

      {/* Subscription status */}
      {subLoading ? (
        <div className="h-36 rounded-2xl bg-muted/40 animate-pulse mb-8" />
      ) : subError ? (
        <div className="rounded-2xl border border-destructive/30 bg-destructive/5 p-6 mb-8">
          <p className="text-sm text-destructive">
            Unable to load subscription information. Please refresh the page.
          </p>
        </div>
      ) : subscription ? (
        <div className="mb-8">
          <SubscriptionStatusCard subscription={subscription} plans={PLANS} />
        </div>
      ) : (
        <div className="rounded-2xl border border-border bg-muted/30 p-6 mb-8">
          <p className="text-sm text-muted-foreground">
            You don&apos;t have an active subscription. Choose a plan below to get started.
          </p>
        </div>
      )}

      {/* Action error (checkout or portal) */}
      {actionError && (
        <div className="mb-4 rounded-lg border border-destructive/30 bg-destructive/5 p-4">
          <p className="text-sm text-destructive">{actionError}</p>
        </div>
      )}

      {/* Billing cycle toggle */}
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-base font-semibold text-foreground">Choose a plan</h2>
        <div className="flex items-center gap-1 rounded-lg border border-border p-1 bg-muted/30">
          <Button
            type="button"
            size="sm"
            variant={billingCycle === 'monthly' ? 'default' : 'ghost'}
            className="h-7 text-xs px-3"
            onClick={() => setBillingCycle('monthly')}
          >
            Monthly
          </Button>
          <Button
            type="button"
            size="sm"
            variant={billingCycle === 'yearly' ? 'default' : 'ghost'}
            className="h-7 text-xs px-3"
            onClick={() => setBillingCycle('yearly')}
          >
            Yearly
            <span className="ml-1 text-xs opacity-70">-17%</span>
          </Button>
        </div>
      </div>

      {/* Plan grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {PLANS.map((plan, idx) => (
          <PlanCard
            key={plan.id}
            plan={plan}
            billingCycle={billingCycle}
            relation={relationFor(
              plan.code,
              currentTierIndex,
              idx,
              subscription?.billing_cycle as BillingCycle | '' | undefined,
              billingCycle
            )}
            isEnterprise={plan.code === 'enterprise'}
            isLoading={actionLoading}
            onChoose={handleChoosePlan}
          />
        ))}
      </div>
    </div>
  )
}
