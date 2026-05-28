'use client'

import { BillingApi, type UpdateSubscriptionDto } from '@workspace/services'
import { useCallback, useState } from 'react'
import type { BillingCycle } from '../domain/subscription'

interface UseUpdateSubscriptionReturn {
  isLoading: boolean
  error: string | null
  preview: UpdateSubscriptionDto | null
  previewUpdate: (planId: string, cycle: BillingCycle) => Promise<UpdateSubscriptionDto | null>
  applyUpdate: (planId: string, cycle: BillingCycle) => Promise<UpdateSubscriptionDto | null>
  clear: () => void
}

/**
 * In-place plan change hook. previewUpdate loads the prorated amount without
 * mutating the subscription; applyUpdate runs the actual change with
 * proration_behavior=always_invoice on the backend so the customer is charged
 * immediately. UI reload happens via the caller after applyUpdate resolves.
 */
export function useUpdateSubscription(): UseUpdateSubscriptionReturn {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [preview, setPreview] = useState<UpdateSubscriptionDto | null>(null)

  const previewUpdate = useCallback(
    async (planId: string, cycle: BillingCycle) => {
      setIsLoading(true)
      setError(null)
      try {
        const res = await BillingApi.updateSubscription(planId, cycle, true)
        setPreview(res)
        return res
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to preview upgrade.')
        return null
      } finally {
        setIsLoading(false)
      }
    },
    []
  )

  const applyUpdate = useCallback(
    async (planId: string, cycle: BillingCycle) => {
      setIsLoading(true)
      setError(null)
      try {
        const res = await BillingApi.updateSubscription(planId, cycle, false)
        return res
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to update subscription.')
        return null
      } finally {
        setIsLoading(false)
      }
    },
    []
  )

  const clear = useCallback(() => {
    setPreview(null)
    setError(null)
  }, [])

  return { isLoading, error, preview, previewUpdate, applyUpdate, clear }
}
