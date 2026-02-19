import { AuthApi } from '@workspace/services'
import { extractTenantFromHostname } from '@workspace/shared-http'
import { env } from '@/lib/env'
import { isRedirectError } from 'next/dist/client/components/redirect-error'
import { headers } from 'next/headers'
import { redirect } from 'next/navigation'
import { AuthProvider } from '@/components/providers/auth-provider'

/**
 * Build the tenant subdomain URL to redirect an already-authenticated user.
 */
function buildTenantRedirectUrl(
  host: string,
  protocol: string,
  tenant: string
): string {
  const hostWithoutPort = host.split(':')[0] || ''
  const port = host.includes(':') ? `:${host.split(':')?.[1] ?? ''}` : ''

  const appDomain = env.NEXT_PUBLIC_APP_DOMAIN
  let baseDomain = appDomain
  if (!baseDomain) {
    if (
      hostWithoutPort === 'localhost' ||
      hostWithoutPort === '127.0.0.1'
    ) {
      baseDomain = 'localhost'
    } else {
      baseDomain = hostWithoutPort.split('.').slice(-2).join('.')
    }
  }

  return `${protocol}//${tenant}.${baseDomain}${port}/`
}

export default async function AuthLayout({ children }: { children: React.ReactNode }) {
  const incomingHeaders = await headers()
  const host = incomingHeaders.get('host') || ''
  const tenant = extractTenantFromHostname(host)

  if (!tenant) {
    // Base domain (e.g. localhost:3000/login) — check if user already has a
    // valid session. If so, redirect to their tenant subdomain.
    try {
      const user = await AuthApi.getCurrentUser()
      if (user.tenant) {
        const protocol = incomingHeaders.get('x-forwarded-proto')
          ? `${incomingHeaders.get('x-forwarded-proto')}:`
          : 'http:'
        const url = buildTenantRedirectUrl(host, protocol, user.tenant)
        redirect(url)
      }
    } catch (error: unknown) {
      // Re-throw Next.js redirect (it uses throw internally)
      if (isRedirectError(error)) {
        throw error
      }
      // Any other error (401, network, etc.) — just show the login page
    }

    return <AuthProvider>{children}</AuthProvider>
  }

  try {
    await AuthApi.getCurrentUser()
    redirect('/')
  } catch (error: unknown) {
    // Re-throw Next.js redirect
    if (isRedirectError(error)) {
      throw error
    }
    // Any other error (401, etc.) — allow auth pages to render
  }

  return <AuthProvider>{children}</AuthProvider>
}
