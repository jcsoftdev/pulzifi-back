'use client'

import { useState } from 'react'
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
}

export function OnboardingWizard({ hasOrg }: OnboardingWizardProps) {
  // Email users (hasOrg=true) start at step 2; OAuth users start at step 1.
  const [step, setStep] = useState<1 | 2>(hasOrg ? 2 : 1)

  // Temporarily hold org values while moving between steps (OAuth path only)
  const [pendingOrgName, setPendingOrgName] = useState('')
  const [pendingSubdomain, setPendingSubdomain] = useState('')

  const { submit, submitProfile, isLoading, error, checkSubdomain, subdomainStatus, subdomainMessage } =
    useOnboarding()

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
      await submit({ org_name: pendingOrgName, subdomain: pendingSubdomain }, profile)
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
    <OnboardingProfileForm
      onSubmit={handleProfileSubmit}
      isLoading={isLoading}
      error={error}
    />
  )
}
