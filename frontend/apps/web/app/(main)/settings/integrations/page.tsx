'use client'

import { IntegrationsApi } from '@workspace/services'
import type { Integration } from '@workspace/services'
import { useEffect, useState } from 'react'
import { useSearchParams } from 'next/navigation'
import { notification } from '@/lib/notification'
import { IntegrationsPanel } from '@/features/integrations/ui/integrations-panel'

export default function IntegrationsPage() {
  const searchParams = useSearchParams()
  const [integrations, setIntegrations] = useState<Integration[]>([])
  const [loading, setLoading] = useState(true)

  // After OAuth callback: ?integration=slack&status=connected
  useEffect(() => {
    const integration = searchParams.get('integration')
    const status = searchParams.get('status')
    if (integration && status === 'connected') {
      notification.success({
        title: 'Integration connected',
        description: `${integration} has been connected successfully.`,
      })
      // Clean up query params from URL
      const url = new URL(window.location.href)
      url.searchParams.delete('integration')
      url.searchParams.delete('status')
      url.searchParams.delete('tenant')
      window.history.replaceState({}, '', url.toString())
    }
  }, [searchParams])

  useEffect(() => {
    IntegrationsApi.list()
      .then(setIntegrations)
      .catch(() => {
        notification.error({
          title: 'Failed to load integrations',
          description: 'Please refresh the page.',
        })
      })
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="px-4 md:px-8 lg:px-24 py-8 max-w-4xl">
        <div className="h-8 w-48 bg-muted rounded animate-pulse mb-4" />
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3, 4, 5, 6].map((n) => (
            <div key={n} className="h-40 bg-muted/40 rounded-2xl animate-pulse" />
          ))}
        </div>
      </div>
    )
  }

  return <IntegrationsPanel integrations={integrations} />
}
