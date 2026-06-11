'use client'

import { BillingApi } from '@workspace/services'
import { Loader2 } from 'lucide-react'
import { useEffect, useState } from 'react'
import {
  clearPendingPlanCookie,
  readPendingPlanCookie,
} from '@/features/auth/application/tenant-session-handoff'
import { useOnboarding } from '../application/use-onboarding'
import type { OnboardingProfileValues } from '../domain/types'
import { OnboardingForm } from './onboarding-form'
import { OnboardingProfileForm } from './onboarding-profile-form'

interface OnboardingWizardProps {
  /**
   * When true the user has no org yet (OAuth path) — show step 1 (org creation)
   * then step 2 (profile questions). When false skip straight to step 2.
   */
  hasOrg: boolean
  /** If present, skip profile form and launch Stripe checkout automatically. */
  planCode?: string
  cycle?: string
}

function CheckoutLauncher({ planCode, cycle }: { planCode: string; cycle: string }) {
  const [error, setError] = useState<string>()

  useEffect(() => {
    async function launch() {
      const plans = await BillingApi.getPlans()
      const plan = plans.find((p) => p.code === planCode)
      if (!plan) {
        clearPendingPlanCookie()
        setError(`Plan "${planCode}" not found. Redirecting to dashboard…`)
        setTimeout(() => window.location.assign('/workspaces'), 2500)
        return
      }
      // Backend expects the plan CODE (resolves Stripe price IDs server-side),
      // not the catalog UUID.
      const { checkoutUrl } = await BillingApi.createCheckoutSession(
        plan.code,
        cycle === 'yearly' ? 'yearly' : 'monthly'
      )
      clearPendingPlanCookie()
      window.location.assign(checkoutUrl)
    }
    launch().catch((err: unknown) => {
      const msg = err instanceof Error ? err.message : 'Failed to launch checkout'
      setError(msg)
    })
  }, [
    planCode,
    cycle,
  ])

  if (error) {
    return (
      <div className="flex flex-col items-center gap-3 py-6 text-center">
        <p className="text-sm text-red-500">{error}</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col items-center gap-3 py-8">
      <Loader2 className="size-8 animate-spin text-[var(--pz-accent)]" />
      <p className="text-sm text-[var(--pz-ink-2)]">Launching secure payment…</p>
    </div>
  )
}

export function OnboardingWizard({ hasOrg, planCode, cycle = 'monthly' }: OnboardingWizardProps) {
  // Email users (hasOrg=true) start at step 2; OAuth users start at step 1.
  const [step, setStep] = useState<1 | 2>(hasOrg ? 2 : 1)

  // Temporarily hold org values while moving between steps (OAuth path only)
  const [pendingOrgName, setPendingOrgName] = useState('')
  const [pendingSubdomain, setPendingSubdomain] = useState('')

  // Cookie fallback: recovers the pricing-page plan when a redirect in the
  // auth chain dropped the ?plan= query param (session-expired bounce, manual
  // login, OAuth). Read client-side after mount to avoid hydration mismatch.
  const [cookiePlan, setCookiePlan] = useState<{
    plan: string
    cycle: string
  } | null>(null)
  useEffect(() => {
    if (planCode) return
    setCookiePlan(readPendingPlanCookie())
  }, [
    planCode,
  ])

  const effectivePlan = planCode ?? cookiePlan?.plan
  const effectiveCycle = planCode ? cycle : (cookiePlan?.cycle ?? 'monthly')

  const {
    submit,
    submitProfile,
    isLoading,
    error,
    checkSubdomain,
    subdomainStatus,
    subdomainMessage,
  } = useOnboarding()

  // Email users with a paid plan: skip profile form, go straight to Stripe.
  // Declared after all hooks so the Rules of Hooks are respected (no conditional hook calls).
  if (hasOrg && effectivePlan) {
    return <CheckoutLauncher planCode={effectivePlan} cycle={effectiveCycle} />
  }

  // Step 1 "next" handler — advance to step 2 (OAuth path)
  const handleStep1Next = (orgName: string, subdomain: string) => {
    setPendingOrgName(orgName)
    setPendingSubdomain(subdomain)
    setStep(2)
  }

  // Step 2 submit handler
  const handleProfileSubmit = async (profile: OnboardingProfileValues) => {
    if (!hasOrg) {
      // OAuth path — create org + save profile atomically
      await submit(
        {
          org_name: pendingOrgName,
          subdomain: pendingSubdomain,
        },
        profile
      )
    } else {
      // Email path — org already exists; just save profile
      await submitProfile(profile)
    }
  }

  if (step === 1) {
    return (
      <OnboardingForm
        onNext={handleStep1Next}
        checkSubdomain={checkSubdomain}
        subdomainStatus={subdomainStatus}
        subdomainMessage={subdomainMessage}
      />
    )
  }

  return (
    <OnboardingProfileForm onSubmit={handleProfileSubmit} isLoading={isLoading} error={error} />
  )
}
