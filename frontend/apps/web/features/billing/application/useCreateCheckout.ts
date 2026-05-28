'use client'

import { BillingApi } from '@workspace/services'
import { useCallback, useState } from 'react'
import type { BillingCycle } from '../domain/subscription'

interface UseCreateCheckoutReturn {
  isLoading: boolean
  error: string | null
  createCheckout: (
    planId: string,
    billingCycle: BillingCycle,
    promotionCode?: string
  ) => Promise<void>
}

/**
 * Hook that initiates a Stripe Checkout session and redirects the user.
 * On success, window.location is set to the Stripe-hosted checkout URL.
 * An optional promotionCode (from the ?promo= apply link) is forwarded so the
 * discount is pre-applied; an invalid code is ignored server-side.
 */
export function useCreateCheckout(): UseCreateCheckoutReturn {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const createCheckout = useCallback(
    async (planId: string, billingCycle: BillingCycle, promotionCode?: string) => {
      setIsLoading(true)
      setError(null)
      try {
        const { checkoutUrl } = await BillingApi.createCheckoutSession(
          planId,
          billingCycle,
          promotionCode
        )
        window.location.href = checkoutUrl
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to start checkout. Please try again.')
        setIsLoading(false)
      }
      // Note: do NOT setIsLoading(false) on success — page is navigating away
    },
    []
  )

  return {
    isLoading,
    error,
    createCheckout,
  }
}
