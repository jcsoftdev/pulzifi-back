import { authCookieOptions } from '@/lib/cookie-options'
import { getPublicOrigin } from '@/lib/public-origin'
import { getBackendOrigin } from '@/lib/server-config'
import type { NextRequest } from 'next/server'
import { NextResponse } from 'next/server'

/**
 * GET /api/auth/logout?redirectTo=/login
 *
 * Used as a cross-subdomain cleanup step: the tenant subdomain redirects the
 * browser here so the main-domain cookies (set during login) are also cleared.
 */
export async function GET(request: NextRequest) {
  const redirectTo = request.nextUrl.searchParams.get('redirectTo') || '/login'
  const { isSecure, cookieDomain, sameSite } = authCookieOptions(request)
  const cookieOpts = {
    path: '/',
    httpOnly: true,
    maxAge: 0,
    secure: isSecure,
    sameSite,
    ...(cookieDomain ? { domain: cookieDomain } : {}),
  }
  const origin = getPublicOrigin(request)
  const response = NextResponse.redirect(new URL(redirectTo, origin))
  response.cookies.set('access_token', '', cookieOpts)
  response.cookies.set('refresh_token', '', cookieOpts)
  response.cookies.set('tenant_hint', '', cookieOpts)

  return response
}

export async function POST(request: NextRequest) {
  try {
    const cookie = request.headers.get('cookie') || ''

    await fetch(`${getBackendOrigin()}/api/v1/auth/logout`, {
      method: 'POST',
      headers: { Cookie: cookie },
      cache: 'no-store',
    })

    const { isSecure, cookieDomain, sameSite } = authCookieOptions(request)
    const cookieOpts = {
      path: '/',
      httpOnly: true,
      maxAge: 0,
      secure: isSecure,
      sameSite,
      ...(cookieDomain ? { domain: cookieDomain } : {}),
    }
    const nextResponse = NextResponse.json({ success: true })

    nextResponse.cookies.set('access_token', '', cookieOpts)
    nextResponse.cookies.set('refresh_token', '', cookieOpts)
    nextResponse.cookies.set('tenant_hint', '', cookieOpts)

    return nextResponse
  } catch (error) {
    console.error('[Logout] Error:', error)
    return NextResponse.json({ success: false }, { status: 500 })
  }
}
