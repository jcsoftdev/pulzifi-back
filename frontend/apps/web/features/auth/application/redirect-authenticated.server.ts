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
 * Build the apex (base-domain) URL, stripping any tenant subdomain. Used to
 * bounce UNauthenticated visitors off tenant subdomains — subdomains require a
 * logged-in user, so anonymous traffic belongs on the apex (e.g. pulzifi.com).
 */
export function buildApexRedirectUrl(
  host: string,
  protocol: string,
  tenant: string,
  path = '/login'
): string {
  if (env.NEXT_PUBLIC_APP_BASE_URL) {
    const base = new URL(env.NEXT_PUBLIC_APP_BASE_URL)
    const portSuffix = base.port ? `:${base.port}` : ''
    return `${base.protocol}//${base.hostname}${portSuffix}${path}`
  }

  const appDomain = env.NEXT_PUBLIC_APP_DOMAIN
  const hostWithoutPort = host.split(':')[0] || ''
  const hostPort = host.includes(':') ? `:${host.split(':')?.[1] ?? ''}` : ''

  let baseDomain: string
  if (appDomain) {
    baseDomain = appDomain
  } else if (hostWithoutPort.toLowerCase().startsWith(`${tenant.toLowerCase()}.`)) {
    // Dev wildcards (tenant.localhost, tenant.lvh.me): drop the tenant label.
    baseDomain = hostWithoutPort.slice(tenant.length + 1)
  } else {
    baseDomain = hostWithoutPort
  }

  return `${protocol}//${baseDomain}${hostPort}${path}`
}

/**
 * Universal subdomain guard: on a tenant subdomain, bounce an UNauthenticated
 * visitor to the apex (subdomains are for logged-in users only). No-op on the
 * apex, or when a session exists — deciding WHERE to send an authenticated user
 * stays with the caller. Use on public/marketing routes that would otherwise
 * render under a tenant host for anonymous traffic.
 */
export async function requireAuthOnSubdomain(params: {
  host: string
  protocol: string
  tenant: string | null
}): Promise<void> {
  const { host, protocol, tenant } = params
  if (!tenant) return
  try {
    await AuthApi.getCurrentUser()
  } catch (error: unknown) {
    if (isRedirectError(error)) throw error
    // Marketing/public route under a tenant host — send anonymous traffic to the
    // apex root (the public site), not the login page.
    redirect(buildApexRedirectUrl(host, protocol, tenant, '/'))
  }
}

/**
 * Redirects an already-authenticated visitor away from auth surfaces
 * (login page / auth layout) to where they belong:
 *
 * - On the base domain: send to their tenant subdomain (onboarding completed)
 *   or to the onboarding wizard (no org / onboarding pending). Falls back to a
 *   `tenant_hint` cookie when the session lookup fails.
 * - On a tenant subdomain: authenticated → the app (`/workspaces`);
 *   unauthenticated → the apex login (subdomains require a logged-in user).
 *
 * Only a no-op on the base domain with no active session — there the caller
 * renders the auth page.
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
    // Unauthenticated on a tenant subdomain. Subdomains are for logged-in users
    // only, so bounce anonymous traffic to the apex login instead of rendering
    // an auth page under the tenant host.
    redirect(buildApexRedirectUrl(host, protocol, tenant))
  }
}
