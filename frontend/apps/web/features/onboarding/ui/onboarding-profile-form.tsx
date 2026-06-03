'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { cn } from '@workspace/ui/lib/utils'
import { useId, useState } from 'react'
import type { OnboardingProfileValues } from '../domain/types'
import { AuthLabel, ErrorBanner } from '@/features/auth/ui/form-atoms'

const baseInput =
  'h-10 w-full rounded-xl border border-[var(--pz-ink)]/10 bg-[var(--pz-page-bg,#f9f9f9)] px-4 text-sm text-[var(--pz-ink)] outline-none transition-[border-color,box-shadow] placeholder:text-[var(--pz-ink)]/30 focus:border-[var(--pz-accent)]/40 focus:bg-white focus:ring-2 focus:ring-[var(--pz-accent)]/15'

const COMPANY_SIZES = ['1-10', '11-50', '51-100', '100+']

const BUSINESS_TYPES = [
  'Marketing agency',
  'Website design and management',
  'SaaS / Software',
  'Real estate',
  'E-commerce',
  'Other',
]

const COMPETITOR_CHALLENGES = [
  "I don't know what they're changing",
  "I don't understand why I'm losing clients",
  'I react too late to their moves',
  "I don't have time to monitor them",
  'Other',
]

interface OnboardingProfileFormProps {
  onSubmit: (values: OnboardingProfileValues) => Promise<void>
  isLoading?: boolean
  error?: string
}

export function OnboardingProfileForm({ onSubmit, isLoading, error }: OnboardingProfileFormProps) {
  const [companySize, setCompanySize] = useState<string | undefined>()
  const [businessType, setBusinessType] = useState<string | undefined>()
  const [challenges, setChallenges] = useState<string[]>([])
  const [websiteUrl, setWebsiteUrl] = useState('')

  const errorId = useId()
  const businessTypeId = useId()
  const websiteId = useId()
  const companySizeGroupId = useId()
  const challengesGroupId = useId()

  const toggleChallenge = (value: string) => {
    setChallenges((prev) =>
      prev.includes(value) ? prev.filter((c) => c !== value) : [...prev, value]
    )
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    await onSubmit({
      company_size: companySize,
      business_type: businessType,
      competitor_challenges: challenges.length > 0 ? challenges : undefined,
      website_url: websiteUrl || undefined,
    })
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-5"
      aria-label="Tell us about yourself"
    >
      {/* Q1 — Company size (single select) */}
      <div className="flex flex-col gap-2">
        <AuthLabel htmlFor={companySizeGroupId}>
          How many people work at your company?
        </AuthLabel>
        <fieldset id={companySizeGroupId} className="flex flex-wrap gap-2 border-0 m-0 p-0">
          <legend className="sr-only">Company size</legend>
          {COMPANY_SIZES.map((size) => (
            <button
              key={size}
              type="button"
              onClick={() => setCompanySize(size === companySize ? undefined : size)}
              className={cn(
                'rounded-xl border px-4 py-1.5 text-sm font-medium transition-[border-color,background-color,color]',
                companySize === size
                  ? 'border-[var(--pz-accent)] bg-[var(--pz-accent)] text-white'
                  : 'border-[var(--pz-ink)]/10 bg-[var(--pz-page-bg,#f9f9f9)] text-[var(--pz-ink)] hover:border-[var(--pz-accent)]/40 hover:bg-[var(--pz-accent)]/5'
              )}
              aria-pressed={companySize === size}
            >
              {size}
            </button>
          ))}
        </fieldset>
      </div>

      {/* Q2 — Business type (dropdown) */}
      <div className="flex flex-col gap-1.5">
        <AuthLabel htmlFor={businessTypeId}>What best describes your business?</AuthLabel>
        <select
          id={businessTypeId}
          value={businessType ?? ''}
          onChange={(e) => setBusinessType(e.target.value || undefined)}
          className={cn(baseInput, 'cursor-pointer appearance-none')}
        >
          <option value="">Select an option</option>
          {BUSINESS_TYPES.map((type) => (
            <option key={type} value={type}>
              {type}
            </option>
          ))}
        </select>
      </div>

      {/* Q3 — Competitor challenges (multi-select) */}
      <div className="flex flex-col gap-2">
        <AuthLabel htmlFor={challengesGroupId}>
          What's your biggest challenge with competitors right now?
        </AuthLabel>
        <fieldset id={challengesGroupId} className="flex flex-col gap-2 border-0 m-0 p-0">
          <legend className="sr-only">Competitor challenges</legend>
          {COMPETITOR_CHALLENGES.map((challenge) => {
            const checked = challenges.includes(challenge)
            return (
              <button
                key={challenge}
                type="button"
                onClick={() => toggleChallenge(challenge)}
                className={cn(
                  'flex w-full items-center gap-3 rounded-xl border px-4 py-2.5 text-left text-sm transition-[border-color,background-color]',
                  checked
                    ? 'border-[var(--pz-accent)] bg-[var(--pz-accent)]/5 text-[var(--pz-ink)]'
                    : 'border-[var(--pz-ink)]/10 bg-[var(--pz-page-bg,#f9f9f9)] text-[var(--pz-ink)] hover:border-[var(--pz-accent)]/30'
                )}
                aria-pressed={checked}
              >
                <span
                  className={cn(
                    'flex size-4 shrink-0 items-center justify-center rounded border-2 transition-colors',
                    checked
                      ? 'border-[var(--pz-accent)] bg-[var(--pz-accent)]'
                      : 'border-[var(--pz-ink)]/20 bg-white'
                  )}
                  aria-hidden="true"
                >
                  {checked && (
                    <svg
                      viewBox="0 0 10 8"
                      fill="none"
                      className="size-2.5"
                      aria-label="Selected"
                      role="img"
                    >
                      <path
                        d="M1 4l2.5 2.5L9 1"
                        stroke="white"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    </svg>
                  )}
                </span>
                {challenge}
              </button>
            )
          })}
        </fieldset>
      </div>

      {/* Q4 — Website URL (optional) */}
      <div className="flex flex-col gap-1.5">
        <AuthLabel htmlFor={websiteId}>Add your website for more context</AuthLabel>
        <input
          id={websiteId}
          type="url"
          value={websiteUrl}
          onChange={(e) => setWebsiteUrl(e.target.value)}
          placeholder="https://yourcompany.com"
          className={baseInput}
        />
        <p className="text-xs text-[var(--pz-ink)]/40">
          We'll use this to personalize your experience. No website yet? Skip for now.
        </p>
      </div>

      {error && <ErrorBanner id={errorId} message={error} />}

      <Button
        type="submit"
        disabled={isLoading}
        className="mt-0.5 h-10 w-full rounded-xl bg-[var(--pz-accent)] text-sm font-semibold shadow-[var(--pz-shadow-accent)] transition-[opacity,box-shadow,transform] hover:opacity-90 hover:shadow-[var(--pz-shadow-accent-lg)] hover:scale-[1.01] active:scale-[0.99]"
      >
        {isLoading ? 'Saving...' : 'Get started'}
      </Button>
    </form>
  )
}
