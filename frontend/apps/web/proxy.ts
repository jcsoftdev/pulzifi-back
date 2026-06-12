// Deep import (not the package barrel) so the edge proxy bundle stays free of
// the axios-based HTTP clients the barrel re-exports.
import { extractTenantFromHostname } from '@workspace/shared-http/tenant-utils'
import { type NextRequest, NextResponse } from 'next/server'

/**
 * Subdomain access gate (Next.js `proxy` convention — the renamed `middleware`).
 *
 * Tenant subdomains (e.g. acme.pulzifi.com) are the authenticated app surface;
 * the public marketing site lives on the apex (pulzifi.com). This proxy is the
 * single chokepoint that keeps anonymous traffic off subdomains: when a request
 * hits a tenant host with no session cookie, it is bounced to the apex before
 * any page renders.
 *
 * It only checks cookie PRESENCE (it cannot validate the token at the edge). A
 * present-but-stale cookie passes through; the page's own getCurrentUser()
 * check then redirects if the session is actually dead. Checking refresh_token
 * (7d) rather than the short-lived access_token avoids bouncing a user whose
 * access token expired but whose session can still refresh.
 */
export function proxy(req: NextRequest): NextResponse {
  const host = req.headers.get('x-forwarded-host') || req.headers.get('host') || ''
  const tenant = extractTenantFromHostname(host)

  // The Payload CMS (admin + data API) is apex-only — it must never be reachable
  // on a tenant subdomain, so no CMS login/session can be established there.
  // Bounce any CMS route on a subdomain to the apex regardless of session cookie.
  const { pathname, search } = req.nextUrl
  const isCmsRoute = pathname.startsWith('/cms-admin') || pathname.startsWith('/cms-api')
  if (tenant && isCmsRoute) {
    return NextResponse.redirect(buildApexUrl(host, req.nextUrl.protocol, `${pathname}${search}`))
  }

  // Legacy blog slug (pre 2026-06-09 normalization) may still be indexed.
  // Lives here instead of next.config redirects(): those match the source
  // case-INSENSITIVELY, so a case-only redirect loops on its own destination.
  if (pathname === '/blog/What-is-competitive-intelligence') {
    const url = req.nextUrl.clone()
    url.pathname = '/blog/what-is-competitive-intelligence'
    return NextResponse.redirect(url, 308)
  }

  // Apex (no tenant) — public site renders normally.
  if (!tenant) {
    return NextResponse.next()
  }

  // A session cookie is present — let the request through; downstream auth
  // (AuthGuard / per-page getCurrentUser) decides what the user actually sees.
  if (req.cookies.has('refresh_token') || req.cookies.has('access_token')) {
    return NextResponse.next()
  }

  // Anonymous on a tenant subdomain — subdomains require auth.
  return NextResponse.redirect(buildApexUrl(host, req.nextUrl.protocol))
}

/**
 * Builds the apex URL by stripping the tenant subdomain from the current host.
 * Mirrors buildApexRedirectUrl on the server, but stays self-contained so the
 * middleware bundle has no server-only imports.
 */
function buildApexUrl(host: string, protocol: string, path = '/'): string {
  const baseUrl = process.env.NEXT_PUBLIC_APP_BASE_URL
  if (baseUrl) {
    const base = new URL(baseUrl)
    const portSuffix = base.port ? `:${base.port}` : ''
    return `${base.protocol}//${base.hostname}${portSuffix}${path}`
  }

  const appDomain = process.env.NEXT_PUBLIC_APP_DOMAIN
  const hostWithoutPort = host.split(':')[0] || ''
  const hostPort = host.includes(':') ? `:${host.split(':')[1] ?? ''}` : ''

  let baseDomain: string
  if (appDomain) {
    baseDomain = appDomain
  } else {
    const tenant = extractTenantFromHostname(host) ?? ''
    baseDomain =
      tenant && hostWithoutPort.toLowerCase().startsWith(`${tenant.toLowerCase()}.`)
        ? hostWithoutPort.slice(tenant.length + 1)
        : hostWithoutPort
  }

  return `${protocol}//${baseDomain}${hostPort}${path}`
}

/**
 * Run on everything except Next internals, API routes, and static assets.
 * API stays out so server-to-server and BFF auth calls are never bounced.
 */
export const config = {
  matcher: [
    '/((?!_next/|api/|favicon.ico|.*\\.[\\w]+$).*)',
  ],
}
