import { peekNonce } from '@/lib/auth-nonce-store'
import { authCookieOptions } from '@/lib/cookie-options'
import type { NextRequest } from 'next/server'
import { NextResponse } from 'next/server'

/**
 * GET /api/auth/set-base-session?nonce=<uuid>&tenant=<name>&returnTo=<url>
 *
 * Called when the user logs in from a tenant subdomain. Runs at the BASE
 * domain (e.g. localhost:3000) to set HttpOnly auth cookies there, so that
 * both the base domain and the tenant subdomain share the session.
 *
 * The nonce is NOT consumed here — it will be consumed by the subsequent
 * tenant callback redirect (the returnTo target).
 *
 * Also sets a non-sensitive `tenant_hint` cookie so the auth layout on the
 * base domain knows which tenant subdomain to redirect to.
 */
export async function GET(request: NextRequest) {
  const nonce = request.nextUrl.searchParams.get('nonce')
  const tenant = request.nextUrl.searchParams.get('tenant')
  const returnTo = request.nextUrl.searchParams.get('returnTo')

  if (!nonce || !returnTo) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  const tokens = peekNonce(nonce)

  if (!tokens) {
    return NextResponse.redirect(new URL('/login?error=SessionExpired', request.url))
  }

  const response = NextResponse.redirect(returnTo)
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

  if (tenant) {
    response.cookies.set('tenant_hint', tenant, {
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
