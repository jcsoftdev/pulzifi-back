import { AppShell } from '@/components/app-shell'
import { AuthGuard } from '@/components/auth-guard'
import { SidebarFeature } from '@/features/sidebar'
import { UsageService } from '@/features/usage/domain/services/usage-service'
import { NotificationService } from '@/features/notifications/domain/services/notification-service'
import { AuthProvider } from '@/components/providers/auth-provider'

export default async function MainLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  // Fetch de datos en el servidor usando domain services
  // Cada feature tiene su propio service en su domain layer
  const checksData = await UsageService.getChecksData()
  const notificationsData = await NotificationService.getNotificationsData()

  return (
    <AuthProvider>
      <AuthGuard>
        <AppShell
          sidebar={<SidebarFeature />}
          checksData={checksData}
          hasNotifications={notificationsData.hasNotifications}
          notificationCount={notificationsData.notificationCount}
        >
          {children}
        </AppShell>
      </AuthGuard>
    </AuthProvider>
  )
}
