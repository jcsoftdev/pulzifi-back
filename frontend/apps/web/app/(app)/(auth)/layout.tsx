export const dynamic = 'force-dynamic'

import { AuthApi } from '@workspace/services'
import { extractTenantFromHostname } from '@workspace/shared-http'
import { isRedirectError } from 'next/dist/client/components/redirect-error'
import { cookies, headers } from 'next/headers'
import { redirect } from 'next/navigation'
import { AuthProvider } from '@/components/providers/auth-provider'
import { env } from '@/lib/env'

/**
 * Build the tenant subdomain URL to redirect an already-authenticated user.
 */
function buildTenantRedirectUrl(host: string, protocol: string, tenant: string): string {
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

export default async function AuthLayout({ children }: { children: React.ReactNode }) {
  const incomingHeaders = await headers()
  // Prefer x-forwarded-host (set by Railway/proxies to the public domain) over
  // the raw host header (which may be the internal service address, e.g. localhost:8080).
  const host = incomingHeaders.get('x-forwarded-host') || incomingHeaders.get('host') || ''
  const tenant = extractTenantFromHostname(host)

  const protocol = (() => {
    const p = incomingHeaders.get('x-forwarded-proto')
    return p ? `${p}:` : 'http:'
  })()

  async function tryRedirectFromUserOrCookie() {
    try {
      const user = await AuthApi.getCurrentUser()
      if (user.tenant) {
        redirect(buildTenantRedirectUrl(host, protocol, user.tenant))
      } else if (user.status === 'approved' && user.tenant == null) {
        // Authenticated but no org yet — send to onboarding
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
  }

  if (!tenant) {
    // Base domain — attempt to redirect an already-authenticated user to their tenant
    await tryRedirectFromUserOrCookie()
    return <AuthProvider>{children}</AuthProvider>
  }

  try {
    await AuthApi.getCurrentUser()
    redirect('/workspaces')
  } catch (error: unknown) {
    if (isRedirectError(error)) throw error
    // Any other error (401, etc.) — allow auth pages to render
  }

  return <AuthProvider>{children}</AuthProvider>
}
