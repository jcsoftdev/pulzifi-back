'use client'

import type { UpdateSubscriptionDto } from '@workspace/services'
import { Button } from '@workspace/ui/components/atoms/button'
import { useId } from 'react'

interface PlanChangeModalProps {
  isOpen: boolean
  targetPlanName: string
  isPreviewLoading: boolean
  preview: UpdateSubscriptionDto | null
  previewError: string | null
  isApplying: boolean
  onConfirm: () => void
  onCancel: () => void
}

/**
 * Confirmation dialog for usage-based plan changes (Model B).
 *
 * Renders the full proration breakdown returned by the backend:
 *   + New plan charge      (priceNew, full price, billing cycle starts today)
 *   - Usage credit          (priceOld × remaining_checks / allowed_checks)
 *   - Account credit        (Stripe customer.balance auto-applied)
 *   = Amount due today      (clamped to 0)
 */
export function PlanChangeModal({
  isOpen,
  targetPlanName,
  isPreviewLoading,
  preview,
  previewError,
  isApplying,
  onConfirm,
  onCancel,
}: PlanChangeModalProps) {
  const titleId = useId()

  if (!isOpen) return null

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm"
      onClick={(e) => {
        if (e.target === e.currentTarget && !isApplying) onCancel()
      }}
    >
      <div className="max-w-md w-full rounded-2xl border border-border bg-card p-8 shadow-xl mx-4">
        <h2 id={titleId} className="text-xl font-semibold text-foreground">
          Confirm change to {targetPlanName}
        </h2>

        <div className="mt-4 space-y-3 text-sm">
          {isPreviewLoading && (
            <p className="text-muted-foreground">Calculating prorated amount…</p>
          )}

          {previewError && !isPreviewLoading && (
            <p className="text-destructive">{previewError}</p>
          )}

          {preview && !isPreviewLoading && (
            <div className="space-y-3">
              <p className="text-muted-foreground">
                We credit the unused checks from your current plan and your Stripe
                balance before charging the new plan&apos;s first cycle today.
              </p>

              <div className="rounded-lg border border-border bg-muted/30 p-3 space-y-2">
                <BreakdownLine
                  label={`${targetPlanName} plan (first cycle)`}
                  amount={preview.new_plan_charge_cents}
                  currency={preview.currency}
                  sign="+"
                />
                <BreakdownLine
                  label={`Usage credit (${preview.checks_remaining}/${preview.checks_allowed} checks remaining)`}
                  amount={preview.usage_credit_cents}
                  currency={preview.currency}
                  sign="-"
                  hideWhenZero
                />
                <BreakdownLine
                  label="Account credit (Stripe balance)"
                  amount={preview.account_credit_cents}
                  currency={preview.currency}
                  sign="-"
                  hideWhenZero
                />
                <div className="border-t border-border pt-2 flex justify-between items-center">
                  <span className="text-xs uppercase tracking-wide text-muted-foreground">
                    Charged today
                  </span>
                  <span className="text-lg font-semibold text-foreground">
                    {formatAmount(preview.amount_due_cents, preview.currency)}
                  </span>
                </div>
              </div>

              <p className="text-xs text-muted-foreground">
                After today, you&apos;ll be billed{' '}
                {formatAmount(preview.new_plan_charge_cents, preview.currency)} per{' '}
                {preview.billing_cycle === 'yearly' ? 'year' : 'month'} starting from
                this date.
              </p>
            </div>
          )}
        </div>

        <div className="mt-6 flex justify-end gap-2">
          <Button variant="outline" onClick={onCancel} disabled={isApplying}>
            Cancel
          </Button>
          <Button
            onClick={onConfirm}
            disabled={isApplying || isPreviewLoading || Boolean(previewError)}
          >
            {isApplying ? 'Applying…' : 'Confirm change'}
          </Button>
        </div>
      </div>
    </div>
  )
}

function BreakdownLine({
  label,
  amount,
  currency,
  sign,
  hideWhenZero,
}: {
  label: string
  amount: number
  currency: string
  sign: '+' | '-'
  hideWhenZero?: boolean
}) {
  if (hideWhenZero && amount === 0) return null
  return (
    <div className="flex justify-between items-baseline text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span className={sign === '-' ? 'text-emerald-600 dark:text-emerald-400' : 'text-foreground'}>
        {sign === '-' ? '−' : '+'}
        {formatAmount(amount, currency)}
      </span>
    </div>
  )
}

function formatAmount(cents: number, currency: string): string {
  const value = cents / 100
  const code = (currency || 'usd').toUpperCase()
  try {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: code,
    }).format(value)
  } catch {
    return `${value.toFixed(2)} ${code}`
  }
}
