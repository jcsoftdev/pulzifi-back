'use client'

import { AuthApi } from '@workspace/services'
import { Button } from '@workspace/ui/components/atoms/button'
import { LogOut } from 'lucide-react'
import { env } from '@/lib/env'

export function LogoutButton() {
  const handleLogout = async () => {
    await AuthApi.logout()

    // Redirect to base domain login so re-login sets cookie on the parent domain
    const host = window.location.host
    const hostname = window.location.hostname
    const protocol = window.location.protocol
    const appDomain = env.NEXT_PUBLIC_APP_DOMAIN

    const port = host.includes(':') ? `:${host.split(':')[1]}` : ''

    let baseDomainHost: string
    if (appDomain) {
      baseDomainHost = `${appDomain}${port}`
    } else if (hostname === 'localhost' || hostname === '127.0.0.1' || hostname.endsWith('.localhost')) {
      baseDomainHost = `localhost${port}`
    } else {
      const parts = hostname.split('.')
      const baseParts = parts.length > 2 ? parts.slice(1) : parts
      baseDomainHost = `${baseParts.join('.')}${port}`
    }

    // Always bounce through the main-domain logout endpoint so its cookies are cleared too.
    window.location.href = `${protocol}//${baseDomainHost}/api/auth/logout?redirectTo=/login`
  }

  return (
    <Button
      variant="ghost"
      className="w-full justify-start gap-3 px-2 py-1.5 h-auto text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-accent"
      onClick={handleLogout}
    >
      <LogOut className="h-4 w-4" />
      <span>Logout</span>
    </Button>
  )
}
