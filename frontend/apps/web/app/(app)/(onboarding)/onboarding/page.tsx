export const dynamic = 'force-dynamic'

import { AuthApi } from '@workspace/services'
import { headers } from 'next/headers'
import { redirect } from 'next/navigation'
import { OnboardingForm } from '@/features/onboarding/ui/onboarding-form'
import { env } from '@/lib/env'

function buildTenantRedirectUrl(host: string, protocol: string, tenant: string): string {
  if (env.NEXT_PUBLIC_APP_BASE_URL) {
    const base = new URL(env.NEXT_PUBLIC_APP_BASE_URL)
    const portSuffix = base.port ? `:${base.port}` : ''
    return `${base.protocol}//${tenant}.${base.hostname}${portSuffix}/`
  }

  const appDomain = env.NEXT_PUBLIC_APP_DOMAIN
  const hostWithoutPort = host.split(':')[0] || ''
  const hostPort = host.includes(':') ? `:${host.split(':')?.[1] ?? ''}` : ''

  let baseDomain: string
  if (appDomain) {
    baseDomain = appDomain
  } else if (hostWithoutPort === 'localhost' || hostWithoutPort === '127.0.0.1') {
    baseDomain = 'localhost'
  } else {
    baseDomain = hostWithoutPort.split('.').slice(-2).join('.')
  }

  const isLocalDomain = baseDomain === 'localhost' || baseDomain === '127.0.0.1'
  const portSuffix = isLocalDomain ? hostPort : ''

  return `${protocol}//${tenant}.${baseDomain}${portSuffix}/`
}

export default async function OnboardingPage() {
  const incomingHeaders = await headers()
  const host = incomingHeaders.get('x-forwarded-host') || incomingHeaders.get('host') || ''

  const protocol = (() => {
    const p = incomingHeaders.get('x-forwarded-proto')
    return p ? `${p}:` : 'http:'
  })()

  try {
    const user = await AuthApi.getCurrentUser()
    // If user already has an org, redirect them to their tenant subdomain.
    // Otherwise render the form regardless of which host they reached us on —
    // the form lets them pick their own subdomain.
    if (user.tenant) {
      redirect(buildTenantRedirectUrl(host, protocol, user.tenant))
    }
  } catch (error: unknown) {
    // Re-throw Next.js internal errors
    if (error && typeof error === 'object' && 'digest' in error) {
      throw error
    }
    // Auth errors are handled by the layout — render the form anyway
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-[var(--pz-page-bg)] px-4 py-8">
      <div className="w-full max-w-[480px] rounded-2xl border border-[var(--pz-card-border)] bg-white p-8 shadow-[var(--pz-card-shadow-rest)]">
        <div className="mb-6">
          <div className="mb-3 inline-flex items-center gap-2 rounded-full bg-[var(--pz-accent)]/10 px-3 py-1">
            <span className="size-1.5 rounded-full bg-[var(--pz-accent)]" />
            <span className="text-xs font-semibold uppercase tracking-widest text-[var(--pz-accent)]">
              Free 14-day trial
            </span>
          </div>
          <h1 className="font-heading text-2xl font-bold leading-tight text-[var(--pz-ink)]">
            Set up your organization
          </h1>
          <p className="mt-1 text-sm text-[var(--pz-ink-2)]">
            Choose a name and subdomain to get started.
          </p>
        </div>
        <OnboardingForm />
      </div>
    </div>
  )
}
