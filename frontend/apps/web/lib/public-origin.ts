import type { NextRequest } from 'next/server'

/**
 * Returns the public-facing origin (e.g. https://pulzifi.com) by reading
 * x-forwarded-host / x-forwarded-proto headers set by the reverse proxy.
 *
 * Falls back to request.nextUrl.origin when the headers are absent (local dev).
 */
export function getPublicOrigin(request: NextRequest): string {
  const forwardedHost = request.headers.get('x-forwarded-host')
  const forwardedProto = request.headers.get('x-forwarded-proto') || 'https'

  if (forwardedHost) {
    return `${forwardedProto}://${forwardedHost}`
  }

  // Fallback: use the host header (some proxies rewrite this instead)
  const host = request.headers.get('host')
  if (host && !host.startsWith('localhost') && !host.startsWith('127.0.0.1')) {
    return `${forwardedProto}://${host}`
  }

  return request.nextUrl.origin
}
