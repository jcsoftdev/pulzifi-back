import { NotificationService } from '@/features/notifications/domain/services/notification-service'
import { NotificationButton } from '@workspace/ui/components/molecules'
import { NotixAnchor } from '@workspace/notix'

export async function NotificationsLoader() {
  const notificationsData = await NotificationService.getNotificationsData()

  return (
    <NotixAnchor
      as={NotificationButton}
      hasNotifications={notificationsData.hasNotifications}
      notificationCount={notificationsData.notificationCount}
      title={`${notificationsData.notificationCount} Notification${notificationsData.notificationCount === 1 ? '' : 's'}`}
      description="You have unread notifications."
      state="info"
      classNames={{
        title: 'text-sm font-semibold text-foreground',
        description: 'text-sm text-muted-foreground',
      }}
    />
  )
}
