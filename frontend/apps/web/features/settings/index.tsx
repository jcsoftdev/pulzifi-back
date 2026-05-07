import { Suspense } from 'react'
import { UpgradeBanner } from './ui/upgrade-banner'
import { SettingsIntegrations } from './ui/settings-integrations'

interface SettingsFeatureProps {
  userEmail?: string
}

export function SettingsFeature({ userEmail: _userEmail }: Readonly<SettingsFeatureProps>) {
  return (
    <div className="px-4 md:px-8 lg:px-16 py-8 max-w-5xl">
      {/* Page header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-foreground">Settings</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Manage your organization settings and integrations.
        </p>
      </div>

      {/* Upgrade banner */}
      <UpgradeBanner />

      {/* Integrations */}
      <div className="mt-10">
        <h2 className="text-base font-semibold text-foreground">Integrations</h2>
        <p className="text-sm text-muted-foreground mt-0.5">
          Connect Slack, email, and other services to receive alerts and notifications.
        </p>
        <div className="mt-4">
          <Suspense
            fallback={
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {[1, 2, 3, 4, 5, 6].map((n) => (
                  <div key={n} className="h-40 bg-muted/40 rounded-2xl animate-pulse" />
                ))}
              </div>
            }
          >
            <SettingsIntegrations />
          </Suspense>
        </div>
      </div>
    </div>
  )
}
