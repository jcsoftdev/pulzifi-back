export const dynamic = 'force-dynamic'

import { Suspense } from 'react'
import { AppShell } from '@/components/app-shell'
import { AuthGuard } from '@/components/auth-guard'
import { NotificationsLoader } from '@/features/notifications/ui/notifications-loader'
import { SidebarFeature } from '@/features/sidebar'
import { SidebarSkeleton } from '@/features/sidebar/ui/sidebar-skeleton'
import { ChecksDataLoader } from '@/features/usage/ui/checks-data-loader'

export default function MainLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <AuthGuard>
      <AppShell
        sidebar={
          <Suspense fallback={<SidebarSkeleton />}>
            <SidebarFeature />
          </Suspense>
        }
        checksSlot={
          <Suspense fallback={<div className="hidden md:block"><div className="h-7 w-44 bg-muted rounded-md animate-pulse" /></div>}>
            <ChecksDataLoader />
          </Suspense>
        }
        notificationsSlot={
          <Suspense fallback={<div className="w-16 h-8 bg-muted rounded-md animate-pulse" />}>
            <NotificationsLoader />
          </Suspense>
        }
      >
        {children}
      </AppShell>
    </AuthGuard>
  )
}
