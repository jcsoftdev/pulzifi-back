import { AppShell } from '@/components/app-shell'
import { SidebarFeature } from '@/features/sidebar'
import { UsageService } from '@/features/usage/domain/services/usage-service'
import { NotificationService } from '@/features/notifications/domain/services/notification-service'

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
    <AppShell
      sidebar={<SidebarFeature />}
      checksData={checksData}
      hasNotifications={notificationsData.hasNotifications}
      notificationCount={notificationsData.notificationCount}
    >
      {children}
    </AppShell>
  )
}
