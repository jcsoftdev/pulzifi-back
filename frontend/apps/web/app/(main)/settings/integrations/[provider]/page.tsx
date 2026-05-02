'use client'

import { DestinationsApi, IntegrationsApi, OrganizationApi } from '@workspace/services'
import type { Destination, Integration } from '@workspace/services'
import { useEffect, useState, use } from 'react'
import { notification } from '@/lib/notification'
import { DestinationForm } from '@/features/integrations/ui/destination-form'
import { DestinationsList } from '@/features/integrations/ui/destinations-list'
import Link from 'next/link'

interface ProviderPageProps {
  params: Promise<{ provider: string }>
}

export default function ProviderPage({ params }: ProviderPageProps) {
  const { provider } = use(params)

  const [orgId, setOrgId] = useState<string | null>(null)
  const [integration, setIntegration] = useState<Integration | null>(null)
  const [destinations, setDestinations] = useState<Destination[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)

  useEffect(() => {
    async function load() {
      try {
        const [org, integrations] = await Promise.all([
          OrganizationApi.getCurrentOrganization(),
          IntegrationsApi.list(),
        ])
        setOrgId(org.id)

        const found = integrations.find(
          (i) => i.serviceType === provider && i.status === 'active'
        )
        setIntegration(found ?? null)

        const dests = await DestinationsApi.list({
          scope_type: 'org',
          scope_id: org.id,
          service_type: provider,
        })
        setDestinations(dests)
      } catch (err) {
        notification.error({
          title: 'Failed to load page',
          description: err instanceof Error ? err.message : 'Please refresh.',
        })
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [provider])

  const handleDisconnect = async () => {
    if (!integration) return
    try {
      await IntegrationsApi.disconnect(integration.id)
      setIntegration(null)
      notification.success({ title: `${provider} disconnected` })
    } catch (err) {
      notification.error({
        title: 'Failed to disconnect',
        description: err instanceof Error ? err.message : 'Please try again.',
      })
    }
  }

  const handleDestinationAdded = (dest: Destination) => {
    setDestinations((prev) => [...prev, dest])
    setShowForm(false)
  }

  const providerLabel =
    provider.charAt(0).toUpperCase() + provider.slice(1).replace(/_/g, ' ')

  if (loading) {
    return (
      <div className="px-4 md:px-8 lg:px-24 py-8 max-w-3xl space-y-4">
        <div className="h-6 w-32 bg-muted rounded animate-pulse" />
        <div className="h-32 bg-muted/40 rounded-2xl animate-pulse" />
      </div>
    )
  }

  return (
    <div className="px-4 md:px-8 lg:px-24 py-8 max-w-3xl space-y-8">
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Link href="/settings/integrations" className="hover:underline">
          Integrations
        </Link>
        <span>/</span>
        <span className="text-foreground font-medium">{providerLabel}</span>
      </div>

      {/* Integration status */}
      <div className="rounded-2xl border border-border bg-card p-5">
        <div className="flex items-center justify-between gap-4">
          <div>
            <h2 className="text-base font-semibold text-foreground">{providerLabel}</h2>
            <p className="text-sm text-muted-foreground mt-0.5">
              {integration ? 'Connected' : 'Not connected'}
            </p>
          </div>
          {integration ? (
            <button
              type="button"
              onClick={handleDisconnect}
              className="text-xs bg-destructive/10 text-destructive hover:bg-destructive/20 px-3 py-1.5 rounded-md transition-colors"
            >
              Disconnect
            </button>
          ) : (
            <button
              type="button"
              onClick={() => {
                window.location.href = IntegrationsApi.startOAuthURL(provider)
              }}
              className="text-xs bg-primary hover:bg-primary/90 text-primary-foreground px-3 py-1.5 rounded-md transition-colors"
            >
              Connect {providerLabel}
            </button>
          )}
        </div>
      </div>

      {/* Destinations */}
      {orgId && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold text-foreground">Destinations</h3>
            {!showForm && (
              <button
                type="button"
                onClick={() => setShowForm(true)}
                className="text-xs bg-muted hover:bg-muted/70 text-foreground px-3 py-1.5 rounded-md transition-colors"
              >
                + Add destination
              </button>
            )}
          </div>

          {showForm && (
            <div className="rounded-2xl border border-border bg-card p-5">
              <h4 className="text-sm font-medium text-foreground mb-4">New destination</h4>
              <DestinationForm
                providerKey={provider}
                integrationId={integration?.id}
                scopeType="org"
                scopeId={orgId}
                onSuccess={handleDestinationAdded}
              />
              <button
                type="button"
                onClick={() => setShowForm(false)}
                className="mt-3 text-xs text-muted-foreground hover:underline"
              >
                Cancel
              </button>
            </div>
          )}

          <DestinationsList
            providerKey={provider}
            integrationId={integration?.id}
            scopeType="org"
            scopeId={orgId}
            destinations={destinations}
            onUpdated={setDestinations}
          />
        </div>
      )}
    </div>
  )
}
