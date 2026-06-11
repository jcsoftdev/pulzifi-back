'use client'

import { BillingApi, type PlanDto } from '@workspace/services'
import { cn } from '@workspace/ui/lib/utils'
import { Check } from 'lucide-react'
import { useSearchParams } from 'next/navigation'
import { Suspense, useEffect, useState } from 'react'
import { setPendingPlanCookie } from '@/features/auth/application/tenant-session-handoff'
import { useRegister } from '@/features/auth/application/use-register'
import { RegisterForm } from '@/features/auth/ui/register-form'

type RegisterFormBlockData = {
  blockType: 'register-form'
  headline?: string
  subheadline?: string
  trialBadge?: string
}

type Props = {
  block: RegisterFormBlockData
}

function formatPrice(cents: number): string {
  return `$${Math.round(cents / 100)}`
}

/** "1. Account → 2. Payment" progress indicator for the paid flow. */
function CheckoutStepper() {
  return (
    <nav className="mb-6 flex items-center gap-3" aria-label="Checkout progress">
      <div className="flex items-center gap-2">
        <span className="flex size-6 items-center justify-center rounded-full bg-[var(--pz-accent)] text-xs font-bold text-white">
          1
        </span>
        <span className="text-sm font-semibold text-[var(--pz-ink)]">Account</span>
      </div>
      <div className="h-px flex-1 bg-[var(--pz-ink)]/10" />
      <div className="flex items-center gap-2 opacity-50">
        <span className="flex size-6 items-center justify-center rounded-full border border-[var(--pz-ink)]/20 text-xs font-bold text-[var(--pz-ink)]">
          2
        </span>
        <span className="text-sm font-medium text-[var(--pz-ink)]">Payment</span>
      </div>
    </nav>
  )
}

/** Selected plan summary — keeps the purchase visible during account creation. */
function PlanSummaryCard({ plan, cycle }: { plan: PlanDto; cycle: string }) {
  const yearly = cycle === 'yearly'
  const cents = yearly ? plan.price_yearly_cents : plan.price_monthly_cents
  const priceLabel = yearly
    ? `${formatPrice(Math.round(cents / 12))}/month`
    : `${formatPrice(cents)}/month`

  return (
    <div className="mb-6 flex items-center justify-between rounded-xl border border-[var(--pz-accent)]/25 bg-[var(--pz-accent)]/5 px-4 py-3">
      <div className="flex items-center gap-3">
        <span className="flex size-8 items-center justify-center rounded-full bg-[var(--pz-accent)]">
          <Check className="size-4 text-white" strokeWidth={3} />
        </span>
        <div>
          <p className="text-sm font-semibold text-[var(--pz-ink)]">{plan.name}</p>
          <p className="text-xs text-[var(--pz-ink-2)]">
            {yearly ? `Billed annually (${formatPrice(cents)}/yr)` : 'Billed monthly'}
          </p>
        </div>
      </div>
      <p className="text-base font-bold text-[var(--pz-ink)]">{priceLabel}</p>
    </div>
  )
}

function RegisterCard({ block }: Props) {
  const searchParams = useSearchParams()
  const planCode = searchParams.get('plan')
  const cycle = searchParams.get('cycle') ?? 'monthly'
  const paidFlow = !!planCode

  const [plan, setPlan] = useState<PlanDto | null>(null)

  useEffect(() => {
    if (!planCode) return
    // Persist the selection immediately so it survives the Google OAuth path,
    // which never goes through the email register submit.
    setPendingPlanCookie(planCode, cycle)
    BillingApi.getPlans()
      .then((plans) => setPlan(plans.find((p) => p.code === planCode) ?? null))
      .catch(() => setPlan(null))
  }, [
    planCode,
    cycle,
  ])

  const { register, isLoading, error, checkSubdomain, subdomainStatus, subdomainMessage } =
    useRegister()

  const headline = paidFlow ? 'Create your account' : (block.headline ?? 'Create your account')
  const subheadline = paidFlow
    ? "You'll complete payment on the next step"
    : (block.subheadline ?? 'No credit card required')
  const badge = paidFlow
    ? `${(planCode ?? '').charAt(0).toUpperCase()}${(planCode ?? '').slice(1)} Plan · ${cycle === 'yearly' ? 'Annual' : 'Monthly'}`
    : (block.trialBadge ?? 'Free 14-day trial')

  return (
    <div
      className={cn(
        'w-full rounded-2xl border border-[var(--pz-card-border)] bg-white p-8 shadow-[var(--pz-card-shadow-rest)] md:p-10',
        paidFlow ? 'max-w-[460px]' : 'max-w-[520px]'
      )}
    >
      {paidFlow && <CheckoutStepper />}

      <div className="mb-6">
        <div className="mb-3 inline-flex items-center gap-2 rounded-full bg-[var(--pz-accent)]/10 px-3 py-1">
          <span className="size-1.5 rounded-full bg-[var(--pz-accent)]" />
          <span className="text-xs font-semibold uppercase tracking-widest text-[var(--pz-accent)]">
            {badge}
          </span>
        </div>
        <h1 className="font-heading text-3xl font-bold leading-tight text-[var(--pz-ink)]">
          {headline}
        </h1>
        <p className="mt-1.5 text-sm text-[var(--pz-ink-2)]">{subheadline}</p>
      </div>

      {paidFlow && plan && <PlanSummaryCard plan={plan} cycle={cycle} />}

      <RegisterForm
        onSubmit={register}
        isLoading={isLoading}
        error={error}
        onSubdomainChange={checkSubdomain}
        subdomainStatus={subdomainStatus}
        subdomainMessage={subdomainMessage}
        paidFlow={paidFlow}
      />
    </div>
  )
}

export function RegisterFormBlock({ block }: Props) {
  return (
    <main className="flex flex-1 items-center justify-center px-4 pb-12 pt-24">
      <Suspense
        fallback={
          <div className="w-full max-w-[520px] rounded-2xl border border-[var(--pz-card-border)] bg-white p-8 shadow-[var(--pz-card-shadow-rest)] md:p-10">
            <div className="mb-3 h-6 w-32 rounded-full bg-[var(--pz-accent)]/10" />
            <div className="h-9 w-64 rounded bg-[var(--pz-ink)]/5" />
          </div>
        }
      >
        <RegisterCard block={block} />
      </Suspense>
    </main>
  )
}
