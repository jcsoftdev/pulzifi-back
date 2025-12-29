'use client'

import { Header } from '@workspace/ui/components/organisms'
import type { ReactNode } from 'react'

export interface AppShellProps {
  children: ReactNode
  sidebar: ReactNode
  checksData: {
    current: number
    max: number
    refillDate: string
  }
  hasNotifications?: boolean
  notificationCount?: number
}

export function AppShell({
  children,
  sidebar,
  checksData,
  hasNotifications = false,
  notificationCount = 0,
}: Readonly<AppShellProps>) {
  return (
    <div className="flex min-h-screen bg-sidebar">
      <div className="sticky top-0 h-screen">{sidebar}</div>
      <div className="flex-1 flex flex-col bg-background">
        <div className="sticky top-0 z-10 bg-background">
          <Header
            checks={checksData}
            hasNotifications={hasNotifications}
            notificationCount={notificationCount}
            onNotificationClick={() => console.log('Notifications clicked')}
          />
        </div>
        {children}
      </div>
    </div>
  )
}
