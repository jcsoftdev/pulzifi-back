import { consumeNonce } from '@/lib/auth-nonce-store'
import { env } from '@/lib/env'
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

  if (!nonce) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  const tokens = consumeNonce(nonce)

  if (!tokens) {
    return NextResponse.redirect(new URL('/login?error=SessionExpired', request.url))
  }

  const redirectTo = request.nextUrl.searchParams.get('redirectTo') || '/'
  const response = NextResponse.redirect(new URL(redirectTo, request.url))
  const isSecure = request.nextUrl.protocol === 'https:'
  const cookieDomain = env.COOKIE_DOMAIN || undefined
  const sameSite = isSecure ? 'none' : 'lax'

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
