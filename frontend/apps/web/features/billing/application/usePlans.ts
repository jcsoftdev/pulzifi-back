'use client'

import type { PlanDto } from '@workspace/services'
import { BillingApi } from '@workspace/services'
import { useCallback, useEffect, useState } from 'react'
import type { Plan } from '../domain/plan'
import { toPlan } from '../domain/plan'

interface UsePlansReturn {
  plans: Plan[]
  isLoading: boolean
  error: string | null
}

export function usePlans(): UsePlansReturn {
  const [plans, setPlans] = useState<Plan[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetch = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const dtos: PlanDto[] = await BillingApi.getPlans()
      setPlans(dtos.map(toPlan))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load plans.')
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    fetch()
  }, [fetch])

  return { plans, isLoading, error }
}
