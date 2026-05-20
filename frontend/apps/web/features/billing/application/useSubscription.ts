'use client'

import type { SubscriptionDto } from '@workspace/services'
import { BillingApi } from '@workspace/services'
import { useCallback, useEffect, useState } from 'react'

interface UseSubscriptionReturn {
  subscription: SubscriptionDto | null
  isLoading: boolean
  error: string | null
  refresh: () => Promise<void>
}

/**
 * Client hook that fetches the current subscription for the authenticated org.
 * Returns null subscription when billing is not enabled or no subscription exists.
 */
export function useSubscription(): UseSubscriptionReturn {
  const [subscription, setSubscription] = useState<SubscriptionDto | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetch = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await BillingApi.getSubscription()
      // FR7: backend returns 200 with billing_status: null when org has no Stripe sub.
      // Treat that as "no subscription" (same as null) to simplify UI logic.
      setSubscription(data?.billing_status !== null ? data : null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load subscription.')
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    fetch()
  }, [
    fetch,
  ])

  return {
    subscription,
    isLoading,
    error,
    refresh: fetch,
  }
}
