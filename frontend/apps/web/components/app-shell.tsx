'use client'

import { useEffect, useState } from 'react'
import { Header } from '@workspace/ui/components/organisms'
import type { BreadcrumbItem } from '@workspace/ui/components/molecules'
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
  breadcrumbs?: BreadcrumbItem[]
}

export function AppShell({
  children,
  sidebar,
  checksData,
  hasNotifications = false,
  notificationCount = 0,
  breadcrumbs: initialBreadcrumbs,
}: Readonly<AppShellProps>) {
  const [breadcrumbs, setBreadcrumbs] = useState<BreadcrumbItem[] | undefined>(initialBreadcrumbs)

  useEffect(() => {
    const handleBreadcrumbUpdate = (event: Event) => {
      const customEvent = event as CustomEvent<{ breadcrumbs: BreadcrumbItem[] }>
      setBreadcrumbs(customEvent.detail.breadcrumbs.length > 0 ? customEvent.detail.breadcrumbs : undefined)
    }

    window.addEventListener('updateBreadcrumbs', handleBreadcrumbUpdate)
    return () => {
      window.removeEventListener('updateBreadcrumbs', handleBreadcrumbUpdate)
    }
  }, [])

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
            breadcrumbs={breadcrumbs}
          />
        </div>
        {children}
      </div>
    </div>
  )
}
