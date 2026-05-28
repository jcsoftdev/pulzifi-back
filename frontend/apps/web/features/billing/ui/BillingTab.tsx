'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { useRouter, useSearchParams } from 'next/navigation'
import { useEffect, useState } from 'react'
import { useCreateCheckout } from '../application/useCreateCheckout'
import { useCustomerPortal } from '../application/useCustomerPortal'
import { useSubscription } from '../application/useSubscription'
import { useUpdateSubscription } from '../application/useUpdateSubscription'
import type { Plan } from '../domain/plan'
import type { BillingCycle } from '../domain/subscription'
import { PlanCard, type PlanRelation } from './PlanCard'
import { PlanChangeModal } from './PlanChangeModal'
import { SubscriptionStatusCard } from './SubscriptionStatusCard'

/**
 * Static plan catalog.
 *
 * Stripe (test mode) is the source of truth for prices: the values below MUST
 * match the `unit_amount` of the prices keyed by `starter_monthly`,
 * `starter_yearly`, `pro_monthly`, `pro_yearly` in the connected Stripe
 * account. If they drift, the checkout page will show a different price than
 * this card. TODO(plan-catalog-api): replace this hardcode with a
 * `GET /api/v1/billing/plans` endpoint that reads `public.plans` so we have a
 * single source of truth.
 *
 * Order matters — index drives upgrade/downgrade detection (`PLANS[i].code` vs
 * `subscription.plan_code`). When fetching plans from an API later, preserve
 * tier order: starter → pro → enterprise.
 */
/**
 * Canonical feature catalog — every plan card shows the SAME list with
 * per-plan inclusion flags so customers can compare side-by-side and see
 * exactly what unlocks at each tier.
 *
 * To add a new feature: append to FEATURES with one column per plan code.
 * Feature ordering matches the visual hierarchy (workspaces → pages →
 * checks → history → integrations → support tier).
 */
/**
 * Quantitative features — each plan exposes a different value (e.g. number of
 * workspaces). Rendered as "Label: value" with a primary ✓ bullet on every
 * card so the customer compares values directly, not crossed-out lines.
 */
const QUANT_FEATURES: ReadonlyArray<{ label: string; values: Record<string, string> }> = [
  {
    label: 'Workspaces',
    values: { trial: '1', starter: '3', pro: 'Unlimited', enterprise: 'Unlimited' },
  },
  {
    label: 'Pages',
    values: { trial: '5', starter: '10', pro: '100', enterprise: 'Custom' },
  },
  {
    label: 'Checks / month',
    values: { trial: '100', starter: '500', pro: '10,000', enterprise: 'Unlimited' },
  },
  {
    label: 'History',
    values: { trial: '7 days', starter: '7 days', pro: '90 days', enterprise: '90+ days' },
  },
  {
    label: 'AI insights / month',
    values: { trial: '5', starter: '50', pro: '250', enterprise: 'Unlimited' },
  },
]

/**
 * Boolean features — rendered with ✓ when included, ✗ when not. One row per
 * capability, same row on every card so the eye can scan across.
 */
const BOOLEAN_FEATURES: ReadonlyArray<{ label: string; in: Record<string, boolean> }> = [
  { label: 'Slack & email alerts', in: { trial: false, starter: false, pro: true, enterprise: true } },
  { label: 'SLA guarantee',        in: { trial: false, starter: false, pro: false, enterprise: true } },
  { label: 'Priority support',     in: { trial: false, starter: false, pro: false, enterprise: true } },
]

function featuresFor(code: string) {
  const quant = QUANT_FEATURES.map((f) => ({
    label: f.label,
    value: f.values[code] ?? '—',
    included: true,
  }))
  const bools = BOOLEAN_FEATURES.map((f) => ({
    label: f.label,
    included: f.in[code] ?? false,
  }))
  return [...quant, ...bools]
}

const PLANS: Plan[] = [
  {
    id: 'trial',
    name: 'Free Trial',
    code: 'trial',
    description: 'Try Pulzifi free for 15 days. No card required.',
    priceMonthly: 0,
    priceYearly: 0,
    features: featuresFor('trial'),
  },
  {
    id: 'starter',
    name: 'Starter',
    code: 'starter',
    description: 'For small teams getting started with monitoring.',
    priceMonthly: 2700,
    priceYearly: 27000,
    features: featuresFor('starter'),
  },
  {
    id: 'pro',
    name: 'Pro',
    code: 'pro',
    description: 'Everything you need to monitor at scale.',
    priceMonthly: 6200,
    priceYearly: 62000,
    features: featuresFor('pro'),
    isPopular: true,
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    code: 'enterprise',
    description: 'Custom limits, SLA, and dedicated support.',
    priceMonthly: 0,
    priceYearly: 0,
    features: featuresFor('enterprise'),
  },
]

const SALES_EMAIL = 'sales@pulzifi.com'

// Tier ordering for upgrade/downgrade detection. Lower = cheaper tier.
// Decoupled from array index so filtering (e.g. hiding Trial once the user
// pays) doesn't shift the upgrade direction.
const TIER_ORDER: Record<string, number> = {
  trial: 0,
  starter: 1,
  pro: 2,
  enterprise: 3,
}

function relationFor(
  thisPlanCode: string,
  currentPlanCode: string | undefined,
  currentCycle: BillingCycle | '' | undefined,
  selectedCycle: BillingCycle
): PlanRelation {
  if (!currentPlanCode) return 'new'
  const thisTier = TIER_ORDER[thisPlanCode] ?? -1
  const currentTier = TIER_ORDER[currentPlanCode] ?? -1
  if (currentTier === -1) return 'new'
  if (thisTier === currentTier) {
    if (currentCycle && currentCycle !== selectedCycle) return 'upgrade'
    return 'current'
  }
  if (thisTier > currentTier) return 'upgrade'
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
  const router = useRouter()
  const searchParams = useSearchParams()
  // ?promo=<code> from the "Try for $1" apply link. Forwarded to checkout so
  // the discount is pre-applied; an invalid code is ignored server-side.
  const promoCode = searchParams.get('promo') ?? undefined
  const {
    subscription,
    isLoading: subLoading,
    error: subError,
    refresh: refreshSubscription,
  } = useSubscription()
  const { isLoading: checkoutLoading, error: checkoutError, createCheckout } = useCreateCheckout()
  const { isLoading: portalLoading, error: portalError, openPortal } = useCustomerPortal()
  const {
    isLoading: updateLoading,
    error: updateError,
    preview,
    previewUpdate,
    applyUpdate,
    clear: clearUpdate,
  } = useUpdateSubscription()
  const [billingCycle, setBillingCycle] = useState<BillingCycle>('monthly')
  const [pendingChange, setPendingChange] = useState<{
    planId: string
    cycle: BillingCycle
    planName: string
  } | null>(null)

  // Sync the cycle toggle with the active subscription's cycle so the user
  // sees their current plan highlighted on the right tab after the
  // subscription loads. Only fires on the initial subscription load — we do
  // NOT override the user's manual toggle once they start exploring plans.
  useEffect(() => {
    if (subscription?.billing_cycle === 'yearly' || subscription?.billing_cycle === 'monthly') {
      setBillingCycle(subscription.billing_cycle)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [subscription?.billing_cycle])

  const currentPlanCode = subscription?.plan_code
  const isOnTrial = currentPlanCode === 'trial'

  const actionLoading = checkoutLoading || portalLoading || updateLoading
  const actionError = checkoutError || portalError || updateError

  const handleChoosePlan = async (planId: string, cycle: BillingCycle, relation: PlanRelation) => {
    const plan = PLANS.find((p) => p.id === planId)
    if (plan?.code === 'enterprise') {
      window.location.href = `mailto:${SALES_EMAIL}?subject=Enterprise plan inquiry`
      return
    }

    const hasStripeCustomer = Boolean(subscription?.stripe_customer_id)
    const hasActiveSub = Boolean(subscription?.stripe_subscription_id)

    // In-place upgrade/downgrade path: customer + active sub already exist →
    // confirm via modal (preview prorated amount) then mutate the existing
    // Stripe Subscription via our API (proration applied immediately) instead
    // of bouncing through Stripe's hosted Portal UI.
    if (hasStripeCustomer && hasActiveSub && (relation === 'upgrade' || relation === 'downgrade')) {
      setPendingChange({ planId, cycle, planName: plan?.name ?? planId })
      previewUpdate(planId, cycle)
      return
    }

    // Customer exists but no active sub yet (rare — orphan/deferred state):
    // fall back to the Stripe Portal so the user can resume in Stripe's UI.
    if (hasStripeCustomer && (relation === 'upgrade' || relation === 'downgrade')) {
      openPortal()
      return
    }

    // First-time paid pick (no customer yet) — must go through Checkout.
    // Forward the ?promo= code so the first-month-$1 discount pre-applies.
    createCheckout(planId, cycle, promoCode)
  }

  const handleConfirmChange = async () => {
    if (!pendingChange) return
    const res = await applyUpdate(pendingChange.planId, pendingChange.cycle)
    if (res) {
      // Refresh client state (subscription card + plan grid) AND the RSC
      // tree (header checks chip + any other server-rendered usage data).
      // router.refresh() invalidates the Next.js fetch cache for this route
      // without doing a full page navigation. Both happen in parallel so
      // the UI catches up in one round-trip.
      await Promise.all([refreshSubscription(), Promise.resolve(router.refresh())])
      setPendingChange(null)
      clearUpdate()
    }
  }

  const handleCancelChange = () => {
    setPendingChange(null)
    clearUpdate()
  }

  // Only the Trial card is conditional; everything else always renders. The
  // grid columns + page width adapt to the actual card count so 3 cards don't
  // leave an empty 4th column (cramped) and 4 cards aren't squeezed.
  const visiblePlans = PLANS.filter((plan) => plan.code !== 'trial' || isOnTrial)
  // Static class literals (Tailwind JIT needs them whole, not interpolated).
  const gridColsClass = visiblePlans.length >= 4 ? 'lg:grid-cols-4' : 'lg:grid-cols-3'
  const containerWidthClass = visiblePlans.length >= 4 ? 'max-w-7xl' : 'max-w-5xl'

  return (
    <div className={`px-4 md:px-8 lg:px-16 py-8 ${containerWidthClass}`}>
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

      {/* Plan grid — Trial card only renders while the user is on it; once a
          paid plan is active it disappears. Columns adapt to the card count. */}
      <div className={`grid grid-cols-1 sm:grid-cols-2 ${gridColsClass} gap-6`}>
        {visiblePlans.map((plan) => {
          const relation = relationFor(
            plan.code,
            currentPlanCode,
            subscription?.billing_cycle as BillingCycle | '' | undefined,
            billingCycle
          )
          return (
            <PlanCard
              key={plan.id}
              plan={plan}
              billingCycle={billingCycle}
              relation={relation}
              isEnterprise={plan.code === 'enterprise'}
              isLoading={actionLoading}
              onChoose={handleChoosePlan}
            />
          )
        })}
      </div>

      <PlanChangeModal
        isOpen={pendingChange !== null}
        targetPlanName={pendingChange?.planName ?? ''}
        isPreviewLoading={updateLoading && preview === null}
        preview={preview}
        previewError={updateError}
        isApplying={updateLoading && preview !== null}
        onConfirm={handleConfirmChange}
        onCancel={handleCancelChange}
      />
    </div>
  )
}
