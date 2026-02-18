'use client'

import { LogOut } from 'lucide-react'
import { Button } from '@workspace/ui/components/atoms/button'

export function LogoutButton() {
  const handleLogout = async () => {
    await fetch('/api/auth/logout', { method: 'POST' })
    window.location.href = '/login'
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
