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

/**
 * GET /api/auth/logout?redirectTo=/login
 *
 * Used as a cross-subdomain cleanup step: the tenant subdomain redirects the
 * browser here so the main-domain cookies (set during login) are also cleared.
 */
export async function GET(request: NextRequest) {
  const redirectTo = request.nextUrl.searchParams.get('redirectTo') || '/login'
  const isSecure = request.nextUrl.protocol === 'https:'
  const cookieDomain = env.COOKIE_DOMAIN || undefined
  const cookieOpts = {
    path: '/',
    httpOnly: true,
    maxAge: 0,
    secure: isSecure,
    sameSite: isSecure ? 'none' as const : 'lax' as const,
    ...(cookieDomain ? { domain: cookieDomain } : {}),
  }
  const response = NextResponse.redirect(new URL(redirectTo, request.url))
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

    const isSecure = request.nextUrl.protocol === 'https:'
    const cookieDomain = env.COOKIE_DOMAIN || undefined
    const cookieOpts = {
      path: '/',
      httpOnly: true,
      maxAge: 0,
      secure: isSecure,
      sameSite: isSecure ? 'none' as const : 'lax' as const,
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
