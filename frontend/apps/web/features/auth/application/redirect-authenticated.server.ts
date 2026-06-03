import { AuthApi } from '@workspace/services'
import { isRedirectError } from 'next/dist/client/components/redirect-error'
import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'
import { env } from '@/lib/env'

/**
 * Build the tenant subdomain URL to redirect an already-authenticated user.
 */
export function buildTenantRedirectUrl(host: string, protocol: string, tenant: string): string {
  // NEXT_PUBLIC_APP_BASE_URL overrides everything — use it when set.
  // Required when running behind a local HTTPS proxy where the host header
  // and browser location.port don't reflect the real port (e.g. proxy on 443
  // forwarding to Next.js on 3000).
  // Example .env.local: NEXT_PUBLIC_APP_BASE_URL=https://localhost:3000
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

  // Preserve port whenever the incoming host has one. Production hosts on
  // standard ports (80/443) never include a port in the Host header, so this
  // stays empty in prod; dev wildcards like lvh.me/nip.io keep their port.
  return `${protocol}//${tenant}.${baseDomain}${hostPort}/`
}

/**
 * Redirects an already-authenticated visitor away from auth surfaces
 * (login page / auth layout) to where they belong:
 *
 * - On the base domain: send to their tenant subdomain (onboarding completed)
 *   or to the onboarding wizard (no org / onboarding pending). Falls back to a
 *   `tenant_hint` cookie when the session lookup fails.
 * - On a tenant subdomain: send to the app (`/workspaces`).
 *
 * No-op when there is no active session — the caller renders the auth page.
 */
export async function redirectAuthenticatedUser(params: {
  host: string
  protocol: string
  tenant: string | null
}): Promise<void> {
  const { host, protocol, tenant } = params

  if (!tenant) {
    // Base domain — attempt to redirect an already-authenticated user to their tenant
    try {
      const user = await AuthApi.getCurrentUser()
      if (user.tenant && user.onboardingCompleted) {
        redirect(buildTenantRedirectUrl(host, protocol, user.tenant))
      } else if (user.tenant && !user.onboardingCompleted) {
        // Has org but hasn't completed onboarding questions — send to wizard (step 2)
        redirect('/onboarding')
      } else if (user.tenant == null) {
        // Authenticated but no org yet — send to onboarding (full wizard)
        redirect('/onboarding')
      }
    } catch (error: unknown) {
      if (isRedirectError(error)) throw error
      const cookieStore = await cookies()
      const tenantHint = cookieStore.get('tenant_hint')?.value
      if (tenantHint) {
        redirect(buildTenantRedirectUrl(host, protocol, tenantHint))
      }
    }
    return
  }

  try {
    await AuthApi.getCurrentUser()
    redirect('/workspaces')
  } catch (error: unknown) {
    if (isRedirectError(error)) throw error
    // Any other error (401, etc.) — allow auth pages to render
  }
}
