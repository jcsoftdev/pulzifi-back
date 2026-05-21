'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import Link from 'next/link'
import { useId } from 'react'
import { useTrialStatus } from '../application/useTrialStatus'

/**
 * Blocking modal shown when the trial has expired and the org has not yet
 * converted to a paid plan. The user can navigate to /billing to upgrade;
 * the modal is intentionally non-dismissible so writes stay gated until
 * the subscription is active.
 *
 * The backend already enforces the gate (HTTP 402 via the trial_guard
 * middleware); the modal is the UX counterpart so users do not see
 * mysterious 402 responses scattered across the UI.
 */
export function UpgradeModal() {
  const { trial, isLoading } = useTrialStatus()
  const titleId = useId()

  if (isLoading || !trial) return null
  if (!trial.is_expired) return null
  if (trial.converted) return null

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm"
    >
      <div className="max-w-md rounded-2xl border border-border bg-card p-8 shadow-xl">
        <h2 id={titleId} className="text-xl font-semibold text-foreground">
          Your trial has ended
        </h2>
        <p className="mt-3 text-sm text-muted-foreground">
          Your dashboards are still readable, but new pages, checks, and integrations are paused
          until you upgrade. Pick a plan to keep monitoring with no interruption.
        </p>
        <div className="mt-6 flex justify-end gap-2">
          <Link href="/billing">
            <Button size="lg">Upgrade now</Button>
          </Link>
        </div>
      </div>
    </div>
  )
}
