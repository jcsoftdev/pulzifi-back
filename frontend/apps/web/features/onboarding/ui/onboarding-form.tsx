'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { cn } from '@workspace/ui/lib/utils'
import { CheckCircle, Loader2, XCircle } from 'lucide-react'
import { useId, useState } from 'react'
import type { SubdomainStatus } from '../application/use-onboarding'
import { AuthLabel, ErrorBanner, FieldMessage } from '@/features/auth/ui/form-atoms'

const baseInput =
  'h-10 w-full rounded-xl border border-[var(--pz-ink)]/10 bg-[var(--pz-page-bg,#f9f9f9)] px-4 text-sm text-[var(--pz-ink)] outline-none transition-[border-color,box-shadow] placeholder:text-[var(--pz-ink)]/30 focus:border-[var(--pz-accent)]/40 focus:bg-white focus:ring-2 focus:ring-[var(--pz-accent)]/15'

function SubdomainStatusIcon({ status }: { status: SubdomainStatus }) {
  if (status === 'checking') {
    return <Loader2 className="size-4 animate-spin text-[var(--pz-ink)]/30" />
  }
  if (status === 'available') {
    return <CheckCircle className="size-4 text-green-500" />
  }
  if (status === 'unavailable') {
    return <XCircle className="size-4 text-destructive" />
  }
  return null
}

interface OnboardingFormProps {
  /** Called when the user completes step 1 — moves to the profile questions step. */
  onNext: (orgName: string, subdomain: string) => void
  checkSubdomain: (subdomain: string) => void
  subdomainStatus: SubdomainStatus
  subdomainMessage?: string
}

export function OnboardingForm({
  onNext,
  checkSubdomain,
  subdomainStatus,
  subdomainMessage,
}: OnboardingFormProps) {
  const [orgName, setOrgName] = useState('')
  const [subdomain, setSubdomain] = useState('')
  const [localError, setLocalError] = useState<string>()

  const orgNameId = useId()
  const subdomainId = useId()
  const subdomainHintId = useId()
  const errorId = useId()

  const handleSubdomainChange = (value: string) => {
    const normalized = value.toLowerCase()
    setSubdomain(normalized)
    checkSubdomain(normalized)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setLocalError(undefined)
    if (!orgName.trim()) {
      setLocalError('Organization name is required.')
      return
    }
    if (!subdomain.trim()) {
      setLocalError('Subdomain is required.')
      return
    }
    if (subdomainStatus === 'unavailable') return
    onNext(orgName, subdomain)
  }

  const isSubdomainUnavailable = subdomainStatus === 'unavailable'

  const subdomainHintVariant = isSubdomainUnavailable
    ? 'error'
    : subdomainStatus === 'available'
      ? 'success'
      : 'hint'

  const subdomainHintText = isSubdomainUnavailable
    ? (subdomainMessage ?? 'Subdomain unavailable')
    : subdomainStatus === 'available'
      ? 'Available'
      : 'Lowercase, numbers, hyphens only'

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-3" aria-label="Set up your organization">
      {/* Organization name */}
      <div className="flex flex-col gap-1.5">
        <AuthLabel htmlFor={orgNameId}>Organization name</AuthLabel>
        <input
          id={orgNameId}
          type="text"
          value={orgName}
          onChange={(e) => setOrgName(e.target.value)}
          required
          aria-required="true"
          placeholder="Acme Inc."
          className={baseInput}
        />
      </div>

      {/* Subdomain */}
      <div className="flex flex-col gap-1.5">
        <AuthLabel htmlFor={subdomainId}>Subdomain</AuthLabel>
        <div className="relative">
          <input
            id={subdomainId}
            type="text"
            value={subdomain}
            onChange={(e) => handleSubdomainChange(e.target.value)}
            required
            aria-required="true"
            aria-invalid={isSubdomainUnavailable}
            aria-describedby={subdomainHintId}
            placeholder="your-company"
            className={cn(
              baseInput,
              'pr-10',
              isSubdomainUnavailable &&
                'border-destructive/50 focus:border-destructive/50 focus:ring-destructive/15',
              subdomainStatus === 'available' &&
                'border-green-500/50 focus:border-green-500/50 focus:ring-green-500/15'
            )}
          />
          <div className="absolute right-3 top-1/2 -translate-y-1/2" aria-hidden="true">
            <SubdomainStatusIcon status={subdomainStatus} />
          </div>
        </div>
        <FieldMessage id={subdomainHintId} variant={subdomainHintVariant}>
          {subdomainStatus === 'checking' ? (
            <span>
              <span className="sr-only">Checking subdomain availability</span>
              <span aria-hidden="true">Lowercase, numbers, hyphens only</span>
            </span>
          ) : (
            subdomainHintText
          )}
        </FieldMessage>
      </div>

      {/* General error */}
      {localError && <ErrorBanner id={errorId} message={localError} />}

      {/* Next */}
      <Button
        type="submit"
        disabled={isSubdomainUnavailable || subdomainStatus === 'checking'}
        className="mt-0.5 h-10 w-full rounded-xl bg-[var(--pz-accent)] text-sm font-semibold shadow-[var(--pz-shadow-accent)] transition-[opacity,box-shadow,transform] hover:opacity-90 hover:shadow-[var(--pz-shadow-accent-lg)] hover:scale-[1.01] active:scale-[0.99]"
      >
        Next
      </Button>
    </form>
  )
}
