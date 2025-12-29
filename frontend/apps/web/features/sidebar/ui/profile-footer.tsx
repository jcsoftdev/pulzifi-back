'use client'

import { LogOut } from 'lucide-react'
import { signOut } from 'next-auth/react'
import { Button } from '@workspace/ui/components/atoms/button'
import { Avatar, AvatarImage, AvatarFallback } from '@workspace/ui/components/atoms/avatar'
import type { User } from '../domain/types'

export interface ProfileFooterProps {
  user: User
}

export function ProfileFooter({ user }: Readonly<ProfileFooterProps>) {
  const handleLogout = async () => {
    await signOut({
      callbackUrl: '/login',
    })
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
        >
          <LogOut className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}
