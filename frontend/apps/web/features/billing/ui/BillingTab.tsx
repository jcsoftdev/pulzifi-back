'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { useRouter, useSearchParams } from 'next/navigation'
import { useEffect, useState } from 'react'
import { usePlans } from '../application/usePlans'
import { useCreateCheckout } from '../application/useCreateCheckout'
import { useCustomerPortal } from '../application/useCustomerPortal'
import { useSubscription } from '../application/useSubscription'
import { useUpdateSubscription } from '../application/useUpdateSubscription'
import type { Plan } from '../domain/plan'
import type { BillingCycle } from '../domain/subscription'
import { PlanCard, type PlanRelation } from './PlanCard'
import { PlanChangeModal } from './PlanChangeModal'
import { SubscriptionStatusCard } from './SubscriptionStatusCard'

const SALES_EMAIL = 'sales@pulzifi.com'

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

export function BillingTab() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const promoCode = searchParams.get('promo') ?? undefined

  const { plans, isLoading: plansLoading, error: plansError } = usePlans()
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
    const plan = plans.find((p) => p.id === planId)
    if (plan?.code === 'enterprise') {
      window.location.href = `mailto:${SALES_EMAIL}?subject=Enterprise plan inquiry`
      return
    }
    const hasStripeCustomer = Boolean(subscription?.stripe_customer_id)
    const hasActiveSub = Boolean(subscription?.stripe_subscription_id)
    if (hasStripeCustomer && hasActiveSub && (relation === 'upgrade' || relation === 'downgrade')) {
      setPendingChange({ planId, cycle, planName: plan?.name ?? planId })
      previewUpdate(planId, cycle)
      return
    }
    if (hasStripeCustomer && (relation === 'upgrade' || relation === 'downgrade')) {
      openPortal()
      return
    }
    createCheckout(planId, cycle, promoCode)
  }

  const handleConfirmChange = async () => {
    if (!pendingChange) return
    const res = await applyUpdate(pendingChange.planId, pendingChange.cycle)
    if (res) {
      await Promise.all([refreshSubscription(), Promise.resolve(router.refresh())])
      setPendingChange(null)
      clearUpdate()
    }
  }

  const handleCancelChange = () => {
    setPendingChange(null)
    clearUpdate()
  }

  const visiblePlans = plans.filter((plan: Plan) => plan.code !== 'trial' || isOnTrial)
  const gridColsClass = visiblePlans.length >= 4 ? 'lg:grid-cols-4' : 'lg:grid-cols-3'
  const containerWidthClass = visiblePlans.length >= 4 ? 'max-w-7xl' : 'max-w-5xl'

  return (
    <div className={`px-4 md:px-8 lg:px-16 py-8 ${containerWidthClass}`}>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-foreground">Billing</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Manage your subscription and payment method.
        </p>
      </div>

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
          <SubscriptionStatusCard subscription={subscription} plans={plans} />
        </div>
      ) : (
        <div className="rounded-2xl border border-border bg-muted/30 p-6 mb-8">
          <p className="text-sm text-muted-foreground">
            You don&apos;t have an active subscription. Choose a plan below to get started.
          </p>
        </div>
      )}

      {actionError && (
        <div className="mb-4 rounded-lg border border-destructive/30 bg-destructive/5 p-4">
          <p className="text-sm text-destructive">{actionError}</p>
        </div>
      )}

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

      {plansError && (
        <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-4 mb-6">
          <p className="text-sm text-destructive">
            Unable to load plan information. Please refresh the page.
          </p>
        </div>
      )}

      <div className={`grid grid-cols-1 sm:grid-cols-2 ${gridColsClass} gap-6`}>
        {plansLoading
          ? Array.from({ length: 3 }, (_, i) => `skeleton-${i}`).map((id) => (
              <div
                key={id}
                className="h-96 rounded-2xl bg-muted/40 animate-pulse"
              />
            ))
          : visiblePlans.map((plan) => {
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
