'use client'

import { Avatar, AvatarFallback, AvatarImage } from '@workspace/ui/components/atoms/avatar'
import { Button } from '@workspace/ui/components/atoms/button'
import { LogOut } from 'lucide-react'
import { useState } from 'react'
import type { User } from '../domain/types'

export interface ProfileFooterProps {
  user: User
}

export function ProfileFooter({ user }: Readonly<ProfileFooterProps>) {
  const [isLoggingOut, setIsLoggingOut] = useState(false)

  const handleLogout = async () => {
    setIsLoggingOut(true)
    await fetch('/api/auth/logout', {
      method: 'POST',
    })

    // Redirect to login without subdomain
    const protocol = globalThis.window?.location.protocol || 'http:'
    const hostname = globalThis.window?.location.hostname || 'localhost'
    const port = globalThis.window?.location.port
    const portStr = port ? `:${port}` : ''

    const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1'
    const isLocalhostSubdomain = hostname.endsWith('.localhost')

    let baseHost: string
    if (isLocalhost || isLocalhostSubdomain) {
      baseHost = `localhost${portStr}`
    } else {
      const parts = hostname.split('.')
      const baseDomain = parts.length > 2 ? parts.slice(1).join('.') : hostname
      baseHost = `${baseDomain}${portStr}`
    }

    // Always bounce through the main-domain logout endpoint so its cookies are cleared too.
    globalThis.window?.location.replace(`${protocol}//${baseHost}/api/auth/logout?redirectTo=/login`)
  }

  return (
    <div className="p-2">
      <div className="flex items-center gap-2 p-2 border border-border rounded-lg bg-card">
        <Avatar className="w-8 h-8 rounded-lg">
          {user.avatar && <AvatarImage src={user.avatar} alt={user.name} />}
          <AvatarFallback className="text-xs rounded-lg">
            {user.name.charAt(0).toUpperCase()}
          </AvatarFallback>
        </Avatar>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-semibold text-foreground truncate leading-tight">
            {user.name}
          </p>
          <p className="text-xs font-normal text-muted-foreground truncate leading-tight">
            {user.role}
          </p>
        </div>
        <Button
          variant="ghost"
          size="icon-sm"
          className="h-4 w-4 flex-shrink-0"
          aria-label="Logout"
          onClick={handleLogout}
          disabled={isLoggingOut}
        >
          <LogOut className={`h-4 w-4 ${isLoggingOut ? 'animate-spin' : ''}`} />
        </Button>
      </div>
    </div>
  )
}
