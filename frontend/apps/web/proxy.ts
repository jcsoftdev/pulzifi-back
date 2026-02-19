import { extractTenantFromHostname } from '@workspace/shared-http'
import type { NextRequest } from 'next/server'
import { NextResponse } from 'next/server'
import { env } from '@/lib/env'

/**
 * Returns the base domain login URL (without tenant subdomain).
 * E.g. on tenant.pulzifi.local:3000 → http://pulzifi.local:3000/login
 */
function getBaseDomainLoginUrl(request: NextRequest, callbackPath?: string): URL {
  const host = request.headers.get('host') || ''
  const protocol = request.nextUrl.protocol // "http:" or "https:"
  const appDomain = env.NEXT_PUBLIC_APP_DOMAIN

  let baseDomainHost: string
  if (appDomain) {
    // Use configured app domain, preserving the port from the current host
    const port = host.includes(':') ? `:${host.split(':')[1]}` : ''
    baseDomainHost = `${appDomain}${port}`
  } else {
    // Fallback: strip the tenant subdomain from the current host
    const hostWithoutPort = host.split(':')[0] || ''
    const port = host.includes(':') ? `:${host.split(':')[1]}` : ''

    if (hostWithoutPort.endsWith('.localhost')) {
      // e.g. acme.localhost → localhost
      baseDomainHost = `localhost${port}`
    } else {
      const parts = hostWithoutPort.split('.')
      // Remove the first part (tenant) if there are enough parts
      const baseParts = parts.length > 2 ? parts.slice(1) : parts
      baseDomainHost = `${baseParts.join('.')}${port}`
    }
  }

  const loginUrl = new URL(`${protocol}//${baseDomainHost}/login`)
  if (callbackPath) {
    loginUrl.searchParams.set('callbackUrl', callbackPath)
  }
  return loginUrl
}

/**
 * Auth Proxy - Handles authentication and tenant validation
 */
export async function proxy(request: NextRequest) {
  const path = request.nextUrl.pathname

  // Public pages that don't require tenant validation
  const publicPaths = [
    '/login',
    '/register',
    '/forgot-password',
    '/reset-password',
    '/lecture-ai',
    '/api/auth/callback',
    '/api/auth/set-base-session',
  ]
  const isPublicPath = publicPaths.some((p) => path.startsWith(p))

  if (isPublicPath) {
    return NextResponse.next()
  }

  const apiBase = env.SERVER_API_URL ?? ''
  const backendOrigin = new URL(apiBase).origin

  const host = request.headers.get('host') || ''
  const tenant = extractTenantFromHostname(host)
  const cookie = request.headers.get('cookie') || ''

  // No tenant subdomain → redirect to base domain login
  if (!tenant) {
    return NextResponse.redirect(getBaseDomainLoginUrl(request, path))
  }

  const headers: Record<string, string> = {
    Cookie: cookie,
    'X-Tenant': tenant,
  }

  const meResponse = await fetch(`${backendOrigin}/api/v1/auth/me`, {
    method: 'GET',
    headers,
    cache: 'no-store',
  })

  console.log({meResponse})

  if (meResponse.status === 401 || meResponse.status === 403) {
    // Redirect to base domain login so cookie is set on the parent domain
    return NextResponse.redirect(getBaseDomainLoginUrl(request, path))
  }

  if (!meResponse.ok) {
    return NextResponse.redirect(getBaseDomainLoginUrl(request))
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    '/((?!api|_next|_nextjs|__nextjs|favicon.ico|login|register|forgot-password|reset-password|lecture-ai).*)',
  ],
}
