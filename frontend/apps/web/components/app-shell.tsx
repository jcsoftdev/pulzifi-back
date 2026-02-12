'use client'

import { useEffect, useState } from 'react'
import { Menu } from 'lucide-react'
import { Header } from '@workspace/ui/components/organisms'
import {
  Button,
  Sheet,
  SheetContent,
  SheetTrigger,
} from '@workspace/ui/components/atoms'
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
      const customEvent = event as CustomEvent<{
        breadcrumbs: BreadcrumbItem[]
      }>
      setBreadcrumbs(
        customEvent.detail.breadcrumbs.length > 0 ? customEvent.detail.breadcrumbs : undefined
      )
    }

    window.addEventListener('updateBreadcrumbs', handleBreadcrumbUpdate)
    return () => {
      window.removeEventListener('updateBreadcrumbs', handleBreadcrumbUpdate)
    }
  }, [])

  return (
    <Sheet>
      <div className="flex min-h-screen bg-sidebar">
        <div className="hidden md:block sticky top-0 h-screen">{sidebar}</div>
        <div className="flex-1 flex flex-col bg-background min-w-0">
          <div className="sticky top-0 z-10 bg-background">
            <Header
              checks={checksData}
              hasNotifications={hasNotifications}
              notificationCount={notificationCount}
              onNotificationClick={() => console.log('Notifications clicked')}
              breadcrumbs={breadcrumbs}
            >
              <SheetTrigger asChild>
                <Button
                  variant="ghost"
                  className="md:hidden -ml-2 mr-2 px-2 h-auto"
                >
                  <Menu className="h-5 w-5" />
                </Button>
              </SheetTrigger>
            </Header>
          </div>
          {children}
        </div>
      </div>
      <SheetContent side="left" className="p-0 w-auto border-none">
        {sidebar}
      </SheetContent>
    </Sheet>
  )
}
