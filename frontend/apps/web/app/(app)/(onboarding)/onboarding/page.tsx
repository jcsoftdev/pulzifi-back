export const dynamic = 'force-dynamic'

import { AuthApi } from '@workspace/services'
import { headers } from 'next/headers'
import { redirect } from 'next/navigation'
import { OnboardingWizard } from '@/features/onboarding/ui/onboarding-wizard'
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

  return `${protocol}//${tenant}.${baseDomain}${hostPort}/`
}

type SearchParams = Promise<{ plan?: string; cycle?: string }>

export default async function OnboardingPage({ searchParams }: { searchParams: SearchParams }) {
  const incomingHeaders = await headers()
  const host = incomingHeaders.get('x-forwarded-host') || incomingHeaders.get('host') || ''

  const protocol = (() => {
    const p = incomingHeaders.get('x-forwarded-proto')
    return p ? `${p}:` : 'http:'
  })()

  const params = await searchParams
  const planCode = params.plan
  const cycle = params.cycle ?? 'monthly'

  let hasOrg = false

  try {
    const user = await AuthApi.getCurrentUser()
    // If user already has an org AND onboarding is complete, redirect to app —
    // unless they have a pending plan selection (go to checkout instead).
    if (user.tenant && user.onboardingCompleted && !planCode) {
      redirect(buildTenantRedirectUrl(host, protocol, user.tenant))
    }
    // hasOrg = true means email-registered user with tenant but onboarding not done.
    hasOrg = user.tenant != null
  } catch (error: unknown) {
    // Re-throw Next.js internal errors
    if (error && typeof error === 'object' && 'digest' in error) {
      throw error
    }
    // Auth errors are handled by the layout — render the form anyway
  }

  const isPaidFlow = hasOrg && !!planCode
  const planLabel = planCode
    ? `${planCode.charAt(0).toUpperCase()}${planCode.slice(1)} Plan`
    : null

  const badge = isPaidFlow ? 'Redirecting to payment' : hasOrg ? 'Quick setup' : 'Free 14-day trial'
  const title = isPaidFlow
    ? 'Almost ready!'
    : hasOrg
      ? 'Almost there!'
      : 'Set up your organization'
  const subtitle = isPaidFlow
    ? `Setting up your ${planLabel} subscription…`
    : hasOrg
      ? 'Tell us a bit about your company so we can personalize your experience.'
      : 'Choose a name and subdomain to get started.'

  return (
    <div className="flex min-h-screen items-center justify-center bg-[var(--pz-page-bg)] px-4 py-8">
      <div className="w-full max-w-[520px] rounded-2xl border border-[var(--pz-card-border)] bg-white p-8 shadow-[var(--pz-card-shadow-rest)]">
        <div className="mb-6">
          <div className="mb-3 inline-flex items-center gap-2 rounded-full bg-[var(--pz-accent)]/10 px-3 py-1">
            <span className="size-1.5 rounded-full bg-[var(--pz-accent)]" />
            <span className="text-xs font-semibold uppercase tracking-widest text-[var(--pz-accent)]">
              {badge}
            </span>
          </div>
          <h1 className="font-heading text-2xl font-bold leading-tight text-[var(--pz-ink)]">
            {title}
          </h1>
          <p className="mt-1 text-sm text-[var(--pz-ink-2)]">{subtitle}</p>
        </div>
        <OnboardingWizard hasOrg={hasOrg} planCode={planCode} cycle={cycle} />
      </div>
    </div>
  )
}
