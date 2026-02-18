import { extractTenantFromHostname } from '@workspace/shared-http'
import type { NextRequest } from 'next/server'
import { NextResponse } from 'next/server'

/**
 * Auth Proxy - Handles authentication and tenant validation
 *
 * Per Next.js best practices, this proxy handles authentication logic
 * while Nginx handles tenant extraction from subdomain.
 *
 * See: https://nextjs.org/docs/messages/middleware-to-proxy
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
  ]
  const isPublicPath = publicPaths.some((p) => path.startsWith(p))

  if (isPublicPath) {
    return NextResponse.next()
  }

  const apiBase =
    process.env.SERVER_API_URL ?? process.env.API_URL ?? process.env.NEXT_PUBLIC_API_URL ?? ''
  const backendOrigin = apiBase ? new URL(apiBase).origin : 'http://localhost:9090'

  const host = request.headers.get('host') || ''
  const tenant = extractTenantFromHostname(host)
  const cookie = request.headers.get('cookie') || ''

  if (!tenant) {
    const loginUrl = new URL('/login', request.url)
    loginUrl.searchParams.set('callbackUrl', request.nextUrl.pathname)
    return NextResponse.redirect(loginUrl)
  }

  const headers: Record<string, string> = {
    Cookie: cookie,
  }
  if (tenant) {
    headers['X-Tenant'] = tenant
  }

  const meResponse = await fetch(`${backendOrigin}/api/v1/auth/me`, {
    method: 'GET',
    headers,
    cache: 'no-store',
  })

  if (meResponse.status === 401 || meResponse.status === 403) {
    const loginUrl = new URL('/login', request.url)
    loginUrl.searchParams.set('callbackUrl', request.nextUrl.pathname)
    return NextResponse.redirect(loginUrl)
  }

  if (!meResponse.ok) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    '/((?!api|_next|_nextjs|__nextjs|favicon.ico|login|register|forgot-password|reset-password|lecture-ai).*)',
  ],
}
