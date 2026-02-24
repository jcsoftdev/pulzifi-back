import { consumeNonce } from '@/lib/auth-nonce-store'
import { authCookieOptions } from '@/lib/cookie-options'
import { getPublicOrigin } from '@/lib/public-origin'
import type { NextRequest } from 'next/server'
import { NextResponse } from 'next/server'

/**
 * GET /api/auth/callback?nonce=<uuid>
 *
 * Called after login when the browser has been redirected to the tenant
 * subdomain (e.g. tenant.localhost:3000). This route runs server-side at
 * that origin, retrieves the JWT tokens stored under the nonce, sets them
 * as HttpOnly cookies scoped to this origin, and redirects to the app root.
 */
export async function GET(request: NextRequest) {
  const nonce = request.nextUrl.searchParams.get('nonce')

  const origin = getPublicOrigin(request)

  if (!nonce) {
    return NextResponse.redirect(new URL('/login', origin))
  }

  const tokens = consumeNonce(nonce)

  if (!tokens) {
    return NextResponse.redirect(new URL('/login?error=SessionExpired', origin))
  }

  const redirectTo = request.nextUrl.searchParams.get('redirectTo') || '/'
  const response = NextResponse.redirect(new URL(redirectTo, origin))
  const { isSecure, cookieDomain, sameSite } = authCookieOptions(request)

  response.cookies.set('access_token', tokens.accessToken, {
    path: '/',
    httpOnly: true,
    secure: isSecure,
    sameSite,
    maxAge: tokens.expiresIn,
    ...(cookieDomain ? { domain: cookieDomain } : {}),
  })

  response.cookies.set('refresh_token', tokens.refreshToken, {
    path: '/',
    httpOnly: true,
    secure: isSecure,
    sameSite,
    maxAge: 7 * 24 * 60 * 60,
    ...(cookieDomain ? { domain: cookieDomain } : {}),
  })

  return response
}
