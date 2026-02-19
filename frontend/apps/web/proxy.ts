import { extractTenantFromHostname } from '@workspace/shared-http'
import type { NextRequest } from 'next/server'
import { NextResponse } from 'next/server'
import { env } from '@/lib/env'

function parseTokenFromSetCookies(setCookieHeaders: string[], cookieName: string): string | null {
  for (const header of setCookieHeaders) {
    const eqIdx = header.indexOf('=')
    if (eqIdx === -1) continue
    const name = header.slice(0, eqIdx).trim()
    const afterEq = header.slice(eqIdx + 1)
    const semiIdx = afterEq.indexOf(';')
    const value = semiIdx === -1 ? afterEq.trim() : afterEq.slice(0, semiIdx).trim()
    if (name === cookieName) return value
  }
  return null
}

function buildUpdatedCookieString(
  existingCookies: string,
  newAccessToken: string,
  newRefreshToken: string | null
): string {
  const cookieMap: Record<string, string> = {}
  for (const part of existingCookies.split(';')) {
    const trimmed = part.trim()
    const eqIdx = trimmed.indexOf('=')
    if (eqIdx === -1) continue
    cookieMap[trimmed.slice(0, eqIdx).trim()] = trimmed.slice(eqIdx + 1).trim()
  }
  cookieMap['access_token'] = newAccessToken
  if (newRefreshToken) cookieMap['refresh_token'] = newRefreshToken
  return Object.entries(cookieMap)
    .map(([k, v]) => `${k}=${v}`)
    .join('; ')
}

async function attemptTokenRefresh(
  request: NextRequest,
  backendOrigin: string,
  cookie: string,
  tenant: string
): Promise<NextResponse | null> {
  try {
    const refreshResponse = await fetch(`${backendOrigin}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: {
        Cookie: cookie,
        'X-Tenant': tenant,
        'Content-Type': 'application/json',
      },
      cache: 'no-store',
    })

    if (!refreshResponse.ok) return null

    const setCookieHeaders = refreshResponse.headers.getSetCookie()
    const newAccessToken = parseTokenFromSetCookies(setCookieHeaders, 'access_token')
    const newRefreshToken = parseTokenFromSetCookies(setCookieHeaders, 'refresh_token')

    if (!newAccessToken) return null

    const updatedCookieStr = buildUpdatedCookieString(cookie, newAccessToken, newRefreshToken)

    const retryResponse = await fetch(`${backendOrigin}/api/v1/auth/me`, {
      method: 'GET',
      headers: { Cookie: updatedCookieStr, 'X-Tenant': tenant },
      cache: 'no-store',
    })

    if (!retryResponse.ok) return null

    const requestHeaders = new Headers(request.headers)
    requestHeaders.set('Cookie', updatedCookieStr)
    const response = NextResponse.next({ request: { headers: requestHeaders } })

    const isSecure = request.nextUrl.protocol === 'https:'
    response.cookies.set('access_token', newAccessToken, {
      path: '/',
      httpOnly: true,
      secure: isSecure,
      sameSite: 'lax',
    })
    if (newRefreshToken) {
      response.cookies.set('refresh_token', newRefreshToken, {
        path: '/',
        httpOnly: true,
        secure: isSecure,
        sameSite: 'lax',
        maxAge: 7 * 24 * 60 * 60,
      })
    }

    return response
  } catch {
    return null
  }
}

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
    // Try to refresh the token before redirecting to login
    const refreshed = await attemptTokenRefresh(request, backendOrigin, cookie, tenant)
    if (refreshed) return refreshed
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
