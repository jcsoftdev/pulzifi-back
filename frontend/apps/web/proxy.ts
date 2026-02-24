import { extractTenantFromHostname } from '@workspace/shared-http/tenant-utils'
import { env } from '@workspace/shared-http/env'
import type { NextRequest } from 'next/server'
import { NextResponse } from 'next/server'

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
  const protocol = request.nextUrl.protocol
  const appDomain = env.NEXT_PUBLIC_APP_DOMAIN

  const hostWithoutPortCheck = host.split(':')[0] || ''
  const isLocalhostRequest = hostWithoutPortCheck === 'localhost' || hostWithoutPortCheck === '127.0.0.1' || hostWithoutPortCheck.endsWith('.localhost')
  const effectiveAppDomain = (appDomain === 'localhost' && !isLocalhostRequest) ? undefined : appDomain

  let baseDomainHost: string
  if (effectiveAppDomain) {
    const port = host.includes(':') ? `:${host.split(':')[1]}` : ''
    baseDomainHost = `${effectiveAppDomain}${port}`
  } else {
    const hostWithoutPort = host.split(':')[0] || ''
    const port = host.includes(':') ? `:${host.split(':')[1]}` : ''

    if (hostWithoutPort.endsWith('.localhost')) {
      baseDomainHost = `localhost${port}`
    } else {
      const parts = hostWithoutPort.split('.')
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

export async function proxy(request: NextRequest) {
  const path = request.nextUrl.pathname

  const publicPaths = [
    '/login',
    '/register',
    '/forgot-password',
    '/reset-password',
    '/lecture-ai',
    '/api/auth/callback',
    '/api/auth/set-base-session',
  ]
  const isPublicPath = path === '/' || publicPaths.some((p) => path.startsWith(p))

  if (isPublicPath) {
    return NextResponse.next()
  }

  const apiBase = env.SERVER_API_URL
  if (!apiBase) {
    console.error('[proxy] SERVER_API_URL is not set — redirecting to login')
    return NextResponse.redirect(getBaseDomainLoginUrl(request))
  }
  const backendOrigin = new URL(apiBase).origin

  const host = request.headers.get('host') || ''
  const tenant = extractTenantFromHostname(host)
  const cookie = request.headers.get('cookie') || ''

  console.log(`[proxy] ${request.method} ${path} | host=${host} tenant=${tenant ?? 'none'} backend=${backendOrigin}`)

  if (!tenant) {
    console.log(`[proxy] No tenant on host=${host}, redirecting to base login`)
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

  console.log(`[proxy] /auth/me → ${meResponse.status} for tenant=${tenant}`)

  if (meResponse.status === 401 || meResponse.status === 403) {
    console.log(`[proxy] Auth failed (${meResponse.status}), attempting refresh for tenant=${tenant}`)
    const refreshed = await attemptTokenRefresh(request, backendOrigin, cookie, tenant)
    if (refreshed) {
      console.log(`[proxy] Token refresh succeeded for tenant=${tenant}`)
      return refreshed
    }
    console.log(`[proxy] Token refresh failed, redirecting to login`)
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
