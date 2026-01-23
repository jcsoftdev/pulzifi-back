import { auth, type ExtendedSession } from '@workspace/auth'
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { extractTenantFromHostname } from '@workspace/shared-http'

/**
 * NextAuth Proxy - Handles authentication and tenant validation
 *
 * Per Next.js best practices, this proxy handles authentication logic
 * while Nginx handles tenant extraction from subdomain.
 *
 * See: https://nextjs.org/docs/messages/middleware-to-proxy
 */
export async function proxy(request: NextRequest) {
  const startTime = performance.now()
  const requestId = Math.random().toString(36).substring(2, 11)
  const path = request.nextUrl.pathname
  const method = request.method

  console.log(`[Proxy-${requestId}] START ${method} ${path}`)

  // Public pages that don't require tenant validation
  const publicPaths = [
    '/login',
    '/register',
    '/forgot-password',
    '/reset-password',
  ]
  const isPublicPath = publicPaths.some((p) => path.startsWith(p))

  try {
    const authStartTime = performance.now()
    const session = (await auth()) as ExtendedSession | null
    const authDuration = performance.now() - authStartTime
    console.log(`[Proxy-${requestId}] Auth completed in ${authDuration.toFixed(2)}ms`)

    // If no session, redirect to login
    if (!session) {
      console.warn(`[Proxy-${requestId}] No session found, redirecting to login`)
      const loginUrl = new URL('/login', request.url)
      // If we are on a subdomain (e.g. jcsoftdev-inc.localhost), we want to stay there
      // instead of redirecting to the base domain
      loginUrl.searchParams.set('callbackUrl', request.nextUrl.pathname)
      return NextResponse.redirect(loginUrl)
    }

    console.log(
      `[Proxy-${requestId}] Session found: user=${session.user?.email}, tenant=${session.tenant}`
    )

    // Only validate tenant for protected pages (not login, register, etc)
    if (!isPublicPath) {
      const hostname = request.headers.get('host') || ''
      const subdomainTenant = extractTenantFromHostname(hostname)

      console.log(
        `[Proxy-${requestId}] Hostname: ${hostname}, Subdomain tenant: ${subdomainTenant}`
      )

      // If subdomain has a tenant (e.g., jcsoftdev-inc.app.local) and it doesn't match session, reject
      // But allow if subdomain is just a generic domain like app.local with no specific tenant
      if (
        subdomainTenant &&
        subdomainTenant !== 'app' &&
        session.tenant &&
        subdomainTenant !== session.tenant
      ) {
        console.warn(
          `[Proxy-${requestId}] Tenant mismatch: session=${session.tenant}, subdomain=${subdomainTenant}`
        )
        const loginUrl = new URL('/login', request.url)
        loginUrl.searchParams.set('error', 'tenant_mismatch')
        return NextResponse.redirect(loginUrl)
      }
    }

    const totalDuration = performance.now() - startTime
    console.log(`[Proxy-${requestId}] SUCCESS ${method} ${path} in ${totalDuration.toFixed(2)}ms`)

    return NextResponse.next()
  } catch (error) {
    const totalDuration = performance.now() - startTime
    console.error(
      `[Proxy-${requestId}] ERROR ${method} ${path} in ${totalDuration.toFixed(2)}ms:`,
      error
    )

    // Return 500 error
    return NextResponse.json(
      {
        error: 'Internal Server Error',
        requestId,
      },
      {
        status: 500,
      }
    )
  }
}

export const config = {
  matcher: [
    '/((?!api|_next/static|_next/image|favicon.ico|login|register|forgot-password|reset-password).*)',
  ],
}
