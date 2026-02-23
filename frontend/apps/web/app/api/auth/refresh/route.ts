import { extractTenantFromHostname } from '@workspace/shared-http'
import { authCookieOptions } from '@/lib/cookie-options'
import { getBackendOrigin } from '@/lib/server-config'
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

export async function POST(request: NextRequest) {
  const cookie = request.headers.get('cookie') ?? ''
  const host = request.headers.get('host') ?? ''
  const tenant = extractTenantFromHostname(host)

  const backendHeaders: Record<string, string> = {
    Cookie: cookie,
    'Content-Type': 'application/json',
  }
  if (tenant) {
    backendHeaders['X-Tenant'] = tenant
  }

  let backendResponse: Response
  try {
    backendResponse = await fetch(`${getBackendOrigin()}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: backendHeaders,
      cache: 'no-store',
    })
  } catch {
    return NextResponse.json({ error: 'Failed to contact backend' }, { status: 502 })
  }

  if (!backendResponse.ok) {
    return NextResponse.json({ error: 'Token refresh failed' }, { status: 401 })
  }

  const setCookieHeaders = backendResponse.headers.getSetCookie()
  const newAccessToken = parseTokenFromSetCookies(setCookieHeaders, 'access_token')
  const newRefreshToken = parseTokenFromSetCookies(setCookieHeaders, 'refresh_token')

  if (!newAccessToken) {
    return NextResponse.json({ error: 'No access token in refresh response' }, { status: 500 })
  }

  let expiresIn = 3600
  try {
    const data = await backendResponse.json()
    if (typeof data.expires_in === 'number') expiresIn = data.expires_in
  } catch {
    // Use default
  }

  const { isSecure, cookieDomain, sameSite } = authCookieOptions(request)
  const response = NextResponse.json({ success: true }, { status: 200 })

  response.cookies.set('access_token', newAccessToken, {
    path: '/',
    httpOnly: true,
    secure: isSecure,
    sameSite,
    maxAge: expiresIn,
    ...(cookieDomain ? { domain: cookieDomain } : {}),
  })

  if (newRefreshToken) {
    response.cookies.set('refresh_token', newRefreshToken, {
      path: '/',
      httpOnly: true,
      secure: isSecure,
      sameSite,
      maxAge: 7 * 24 * 60 * 60,
      ...(cookieDomain ? { domain: cookieDomain } : {}),
    })
  }

  return response
}
