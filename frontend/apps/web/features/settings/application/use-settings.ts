'use client'

import { IntegrationApi, type Integration } from '@workspace/services'
import { useCallback, useEffect, useState } from 'react'
import type { ServiceType } from '../domain/types'

export function useSettings() {
  const [integrations, setIntegrations] = useState<Integration[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchIntegrations = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await IntegrationApi.listIntegrations()
      setIntegrations(data)
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to load integrations'))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchIntegrations()
  }, [fetchIntegrations])

  const connectIntegration = useCallback(
    async (serviceType: ServiceType, url: string) => {
      const result = await IntegrationApi.upsertIntegration({
        service_type: serviceType,
        config: { url },
      })
      setIntegrations((prev) => {
        const exists = prev.find((i) => i.serviceType === serviceType)
        if (exists) {
          return prev.map((i) => (i.serviceType === serviceType ? result : i))
        }
        return [...prev, result]
      })
      return result
    },
    []
  )

  const disconnectIntegration = useCallback(async (id: string) => {
    await IntegrationApi.deleteIntegration(id)
    setIntegrations((prev) => prev.filter((i) => i.id !== id))
  }, [])

  const getIntegrationByType = useCallback(
    (serviceType: ServiceType): Integration | undefined => {
      return integrations.find((i) => i.serviceType === serviceType)
    },
    [integrations]
  )

  return {
    integrations,
    loading,
    error,
    connectIntegration,
    disconnectIntegration,
    getIntegrationByType,
    refresh: fetchIntegrations,
  }
}
