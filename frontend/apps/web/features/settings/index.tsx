import Link from 'next/link'
import { UpgradeBanner } from './ui/upgrade-banner'

interface SettingsFeatureProps {
  userEmail?: string
}

export function SettingsFeature({ userEmail: _userEmail }: Readonly<SettingsFeatureProps>) {
  return (
    <div className="px-4 md:px-8 lg:px-24 py-8 max-w-3xl">
      {/* Page header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-foreground">Settings</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Manage your organization settings and integrations.
        </p>
      </div>

      {/* Upgrade banner */}
      <UpgradeBanner />

      {/* Integrations link */}
      <div className="mt-10">
        <h2 className="text-base font-semibold text-foreground">Integrations</h2>
        <p className="text-sm text-muted-foreground mt-0.5">
          Connect Slack, email, and other services to receive alerts and notifications.
        </p>
        <Link
          href="/settings/integrations"
          className="mt-4 inline-flex items-center gap-2 text-sm font-medium text-primary hover:underline"
        >
          Manage integrations &rarr;
        </Link>
      </div>
    </div>
  )
}
