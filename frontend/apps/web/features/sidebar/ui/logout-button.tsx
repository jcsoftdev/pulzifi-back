'use client'

import { AuthApi } from '@workspace/services'
import { Button } from '@workspace/ui/components/atoms/button'
import { LogOut } from 'lucide-react'
import { env } from '@/lib/env'
import { useState } from 'react'

export function LogoutButton() {
  const [isLoggingOut, setIsLoggingOut] = useState(false)

  const handleLogout = async () => {
    if (isLoggingOut) return
    setIsLoggingOut(true)

    // Best-effort: clear subdomain cookies via the BFF. Even if this fails,
    // the main-domain redirect below will still clear those cookies.
    try {
      await AuthApi.logout()
    } catch {
      // continue to cross-domain logout regardless
    }

    const hostname = globalThis.location.hostname
    const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1' || hostname.endsWith('.localhost')
    const appDomain = env.NEXT_PUBLIC_APP_DOMAIN
    const appBaseUrl = env.NEXT_PUBLIC_APP_BASE_URL

    // NEXT_PUBLIC_APP_BASE_URL is the explicit frontend base URL (set in Railway/production).
    // Use it only on localhost so stale build-time values never override a real production URL.
    const base = (isLocalhost && appBaseUrl) ? new URL(appBaseUrl) : null

    // Protocol: prefer NEXT_PUBLIC_APP_BASE_URL when on localhost (e.g. SSH tunnel)
    const protocol = base ? base.protocol : globalThis.location.protocol

    let baseDomainHost: string
    if (appDomain && !(appDomain === 'localhost' && !isLocalhost)) {
      // appDomain explicitly set — use it without the browser's port.
      // Port comes from NEXT_PUBLIC_APP_BASE_URL if set, otherwise omitted (standard HTTPS).
      const domainPort = base?.port ? `:${base.port}` : ''
      baseDomainHost = `${appDomain}${domainPort}`
    } else if (base) {
      // On localhost with NEXT_PUBLIC_APP_BASE_URL — derive base domain and port from it.
      const basePort = base.port ? `:${base.port}` : ''
      const parts = base.hostname.split('.')
      const baseParts = parts.length > 2 ? parts.slice(-2) : parts
      baseDomainHost = `${baseParts.join('.')}${basePort}`
    } else {
      // Derive everything from the current browser URL.
      const port = globalThis.location.port ? `:${globalThis.location.port}` : ''
      if (isLocalhost) {
        baseDomainHost = `localhost${port}`
      } else {
        const parts = hostname.split('.')
        const baseParts = parts.length > 2 ? parts.slice(1) : parts
        baseDomainHost = `${baseParts.join('.')}${port}`
      }
    }

    // Always bounce through the main-domain logout endpoint so its cookies
    // (access_token, refresh_token, tenant_hint) are cleared too.
    globalThis.location.href = `${protocol}//${baseDomainHost}/api/auth/logout?redirectTo=/login`
  }

  return (
    <Button
      variant="ghost"
      className="w-full justify-start gap-3 px-2 py-1.5 h-auto text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-accent"
      onClick={handleLogout}
      disabled={isLoggingOut}
    >
      <LogOut className="h-4 w-4" />
      <span>{isLoggingOut ? 'Signing out...' : 'Logout'}</span>
    </Button>
  )
}
