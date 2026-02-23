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
    const protocol = globalThis.location.protocol
    const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1' || hostname.endsWith('.localhost')
    const appDomain = env.NEXT_PUBLIC_APP_DOMAIN

    const port = globalThis.location.port ? `:${globalThis.location.port}` : ''

    let baseDomainHost: string
    // Ignore NEXT_PUBLIC_APP_DOMAIN=localhost when not actually on localhost
    // (prevents stale build-time value from sending users to localhost in production)
    if (appDomain && !(appDomain === 'localhost' && !isLocalhost)) {
      baseDomainHost = `${appDomain}${port}`
    } else if (isLocalhost) {
      baseDomainHost = `localhost${port}`
    } else {
      const parts = hostname.split('.')
      const baseParts = parts.length > 2 ? parts.slice(1) : parts
      baseDomainHost = `${baseParts.join('.')}${port}`
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
