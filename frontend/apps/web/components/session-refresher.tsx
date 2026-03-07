'use client'

import { useEffect, useRef } from 'react'

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

    fetch('/api/auth/refresh', { method: 'POST', credentials: 'include' })
      .then((res) => {
        if (res.ok) {
          // Full reload ensures the layout (AuthGuard) re-runs with fresh cookies
          window.location.reload()
        } else {
          window.location.href = '/login'
        }
      })
      .catch(() => {
        window.location.href = '/login'
      })
  }, [])

  return null
}
