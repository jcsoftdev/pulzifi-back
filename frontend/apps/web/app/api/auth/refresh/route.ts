import type { NextRequest } from 'next/server'
import { NextResponse } from 'next/server'
import { env } from '@/lib/env'

function getBackendOrigin(): string {
  const apiBase = env.SERVER_API_URL ?? 'http://localhost:9090'
  try {
    return new URL(apiBase).origin
  } catch {
    return 'http://localhost:9090'
  }
}

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

  let backendResponse: Response
  try {
    backendResponse = await fetch(`${getBackendOrigin()}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: {
        Cookie: cookie,
        'Content-Type': 'application/json',
      },
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

  const isSecure = request.nextUrl.protocol === 'https:'
  const cookieDomain = env.COOKIE_DOMAIN || undefined
  const sameSite = isSecure ? 'none' : 'lax'
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
