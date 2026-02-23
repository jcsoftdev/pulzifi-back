import { env } from '@/lib/env'
import type { NextRequest } from 'next/server'

/**
 * Returns shared cookie options for auth tokens.
 *
 * Browsers reject Domain=localhost (it's a public suffix), so we omit the
 * domain attribute for localhost/127.0.0.1 requests — cookies are scoped to
 * the exact hostname that set them. In production COOKIE_DOMAIN (e.g. .example.com)
 * enables sharing across subdomains.
 *
 * isSecure is derived from x-forwarded-proto first so it works correctly
 * behind Railway's TLS-terminating proxy.
 */
export function authCookieOptions(request: NextRequest) {
  const forwardedProto = request.headers.get('x-forwarded-proto')
  const isSecure = forwardedProto ? forwardedProto === 'https' : request.nextUrl.protocol === 'https:'

  const hostname = request.nextUrl.hostname
  const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1' || hostname.endsWith('.localhost')

  const cookieDomain = (!isLocalhost && env.COOKIE_DOMAIN) ? env.COOKIE_DOMAIN : undefined
  const sameSite = isSecure ? 'none' : 'lax'

  return { isSecure, cookieDomain, sameSite } as const
}
