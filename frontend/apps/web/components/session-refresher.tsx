'use client'

import { getTenantFromWindow } from '@workspace/shared-http'
import { useEffect, useRef } from 'react'

/**
 * Builds the apex (base-domain) login URL by stripping the current tenant
 * subdomain. Subdomains require an authenticated user, so a failed refresh on
 * a tenant host sends the visitor to the apex login (e.g. pulzifi.com/login).
 */
function apexLoginUrl(): string {
  const baseUrl = process.env.NEXT_PUBLIC_APP_BASE_URL
  if (baseUrl) {
    const base = new URL(baseUrl)
    const portSuffix = base.port ? `:${base.port}` : ''
    return `${base.protocol}//${base.hostname}${portSuffix}/login`
  }

  const appDomain = process.env.NEXT_PUBLIC_APP_DOMAIN
  if (appDomain) {
    return `https://${appDomain}/login`
  }

  // Dev fallback: drop the tenant label from the current host.
  const { protocol, hostname, port } = window.location
  const tenant = getTenantFromWindow()
  const baseDomain =
    tenant && hostname.toLowerCase().startsWith(`${tenant.toLowerCase()}.`)
      ? hostname.slice(tenant.length + 1)
      : hostname
  const portSuffix = port ? `:${port}` : ''
  return `${protocol}//${baseDomain}${portSuffix}/login`
}

/**
 * Client component that attempts to refresh the session when the server-side
 * auth check fails (access token expired). Since the browser still holds a
 * valid refresh_token cookie, the refresh call works here — unlike on the
 * server where Set-Cookie headers cannot be propagated back to the browser.
 *
 * Uses window.location.reload() instead of router.refresh() because the
 * AuthGuard lives in a layout and router.refresh() may not reliably
 * re-execute layout server components in all Next.js versions.
 */
export function SessionRefresher() {
  const attempted = useRef(false)

  useEffect(() => {
    if (attempted.current) return
    attempted.current = true

    fetch('/api/auth/refresh', {
      method: 'POST',
      credentials: 'include',
    })
      .then((res) => {
        if (res.ok) {
          // Full reload ensures the layout (AuthGuard) re-runs with fresh cookies
          window.location.reload()
        } else {
          window.location.href = apexLoginUrl()
        }
      })
      .catch(() => {
        window.location.href = apexLoginUrl()
      })
  }, [])

  return null
}
