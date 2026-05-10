'use client'

import { AuthApi, IntegrationsApi } from '@workspace/services'
import type { Integration, User } from '@workspace/services'
import { useEffect, useState } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { notification } from '@/lib/notification'
import { IntegrationsPanel } from '@/features/integrations/ui/integrations-panel'
import { TeamsConsentModal } from '@/features/integrations/ui/teams-consent-modal'

export function SettingsIntegrations() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const [integrations, setIntegrations] = useState<Integration[]>([])
  const [me, setMe] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [consentModal, setConsentModal] = useState<{ open: boolean; url: string }>({
    open: false,
    url: '',
  })

  useEffect(() => {
    const integration = searchParams.get('integration')
    const status = searchParams.get('status')
    const integrationError = searchParams.get('integration_error')
    const adminUrl = searchParams.get('admin_url')

    if (integrationError === 'consent_required') {
      setConsentModal({ open: true, url: adminUrl ?? '' })
      router.replace('/settings')
      return
    }

    if (integration && status === 'connected') {
      notification.success({
        title: 'Integration connected',
        description: `${integration} has been connected successfully.`,
      })
      const url = new URL(window.location.href)
      url.searchParams.delete('integration')
      url.searchParams.delete('status')
      url.searchParams.delete('tenant')
      window.history.replaceState({}, '', url.toString())
    }
  }, [searchParams, router])

  useEffect(() => {
    Promise.all([IntegrationsApi.list(), AuthApi.getCurrentUser()])
      .then(([integrationList, user]) => {
        setIntegrations(integrationList)
        setMe(user)
      })
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
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">
        {[1, 2, 3, 4, 5, 6].map((n) => (
          <div key={n} className="h-40 bg-muted/40 rounded-2xl animate-pulse" />
        ))}
      </div>
    )
  }

  return (
    <>
      <IntegrationsPanel integrations={integrations} me={me} />
      {consentModal.open && (
        <TeamsConsentModal
          open={consentModal.open}
          adminURL={consentModal.url}
          onClose={() => setConsentModal({ open: false, url: '' })}
        />
      )}
    </>
  )
}
