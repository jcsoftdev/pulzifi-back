'use client'

import { BillingApi, type CancelSubscriptionDto } from '@workspace/services'
import { useCallback, useState } from 'react'

interface UseCancelSubscriptionReturn {
  isLoading: boolean
  error: string | null
  cancel: () => Promise<CancelSubscriptionDto | null>
  resume: () => Promise<CancelSubscriptionDto | null>
}

/**
 * Cancel / resume the current subscription. Cancellation is scheduled for the
 * end of the paid period — the org keeps access until then, then drops to no
 * active plan. The caller reloads subscription state after the call resolves.
 */
export function useCancelSubscription(): UseCancelSubscriptionReturn {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const cancel = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      return await BillingApi.cancelSubscription()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to cancel subscription.')
      return null
    } finally {
      setIsLoading(false)
    }
  }, [])

  const resume = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      return await BillingApi.resumeSubscription()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to resume subscription.')
      return null
    } finally {
      setIsLoading(false)
    }
  }, [])

  return { isLoading, error, cancel, resume }
}
