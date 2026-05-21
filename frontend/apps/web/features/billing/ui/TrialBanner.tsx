'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import Link from 'next/link'
import { useTrialStatus } from '../application/useTrialStatus'

/**
 * Top-of-page banner shown above the main content while the org is on an
 * active trial. Disappears once the trial converts or the org is on a paid plan.
 *
 * Surfaces "n days left in your trial — Upgrade" with a CTA to /billing.
 * Renders nothing when:
 *   - the trial status couldn't be loaded
 *   - the org isn't on a trial
 *   - the org has already converted
 *   - the trial has fully expired (the UpgradeModal handles that case)
 */
export function TrialBanner() {
  const { trial, isLoading } = useTrialStatus()

  if (isLoading || !trial) return null
  if (!trial.is_trial) return null
  if (trial.converted) return null
  if (trial.is_expired) return null

  const days = trial.days_remaining
  const urgent = days <= 3

  return (
    <output
      className={
        urgent
          ? 'flex items-center justify-between gap-4 px-4 py-2 bg-destructive/10 border-b border-destructive/30 text-sm w-full'
          : 'flex items-center justify-between gap-4 px-4 py-2 bg-primary/10 border-b border-primary/20 text-sm w-full'
      }
    >
      <span className="text-foreground">
        <strong>{days}</strong> {days === 1 ? 'day' : 'days'} left in your trial
        {urgent && ' — add a payment method to keep monitoring'}
      </span>
      <Link href="/billing">
        <Button size="sm" variant={urgent ? 'default' : 'outline'}>
          Upgrade
        </Button>
      </Link>
    </output>
  )
}
