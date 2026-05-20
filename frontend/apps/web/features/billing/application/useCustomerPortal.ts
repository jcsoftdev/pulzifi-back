'use client'

import { BillingApi } from '@workspace/services'
import { useCallback, useState } from 'react'

interface UseCustomerPortalReturn {
  isLoading: boolean
  error: string | null
  openPortal: () => Promise<void>
}

/**
 * Hook that opens the Stripe Customer Portal and redirects the user.
 * On success, window.location is set to the Stripe-hosted portal URL.
 */
export function useCustomerPortal(): UseCustomerPortalReturn {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const openPortal = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const { portalUrl } = await BillingApi.openCustomerPortal()
      window.location.href = portalUrl
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to open billing portal. Please try again.'
      )
      setIsLoading(false)
    }
    // Note: do NOT setIsLoading(false) on success — page is navigating away
  }, [])

  return {
    isLoading,
    error,
    openPortal,
  }
}
