'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { notification } from '@/lib/notification'
import { useCancelSubscription } from '../application/useCancelSubscription'

interface CancelPlanButtonProps {
  /** True when a period-end cancellation is already scheduled. */
  cancelScheduled: boolean
}

/**
 * Always-visible cancel / resume control. When no cancellation is scheduled it
 * offers "Cancel plan" (with confirm). When one is pending it offers "Resume
 * plan" to revert it. Cancellation takes effect at the end of the paid period.
 */
export function CancelPlanButton({ cancelScheduled }: CancelPlanButtonProps) {
  const router = useRouter()
  const { isLoading, error, cancel, resume } = useCancelSubscription()
  const [confirming, setConfirming] = useState(false)

  const handleCancel = async () => {
    const res = await cancel()
    if (res) {
      notification.success({
        title: 'Plan cancellation scheduled',
        description: res.cancel_at
          ? `Your plan stays active until ${new Date(res.cancel_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}.`
          : 'Your plan will be canceled at the end of the current period.',
      })
      setConfirming(false)
      router.refresh()
    }
  }

  const handleResume = async () => {
    const res = await resume()
    if (res) {
      notification.success({
        title: 'Cancellation reverted',
        description: 'Your subscription will renew normally.',
      })
      router.refresh()
    }
  }

  if (cancelScheduled) {
    return (
      <div className="flex flex-col items-start gap-1">
        <Button type="button" variant="outline" size="sm" disabled={isLoading} onClick={handleResume}>
          {isLoading ? 'Resuming...' : 'Resume plan'}
        </Button>
        {error && <p className="text-xs text-destructive">{error}</p>}
      </div>
    )
  }

  if (confirming) {
    return (
      <div className="flex flex-col items-start gap-2">
        <p className="text-xs text-muted-foreground">
          You'll keep access until the end of the current billing period, then your plan ends.
        </p>
        <div className="flex gap-2">
          <Button
            type="button"
            variant="destructive"
            size="sm"
            disabled={isLoading}
            onClick={handleCancel}
          >
            {isLoading ? 'Canceling...' : 'Confirm cancellation'}
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            disabled={isLoading}
            onClick={() => setConfirming(false)}
          >
            Keep plan
          </Button>
        </div>
        {error && <p className="text-xs text-destructive">{error}</p>}
      </div>
    )
  }

  return (
    <div className="flex flex-col items-start gap-1">
      <Button
        type="button"
        variant="ghost"
        size="sm"
        className="text-destructive hover:text-destructive"
        disabled={isLoading}
        onClick={() => setConfirming(true)}
      >
        Cancel plan
      </Button>
    </div>
  )
}
