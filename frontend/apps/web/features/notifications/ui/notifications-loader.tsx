import { NotificationService } from '@/features/notifications/domain/services/notification-service'
import { NotificationsWidget } from './notifications-widget'

export async function NotificationsLoader() {
  const [notificationsData, { notifications, total }] = await Promise.all([
    NotificationService.getNotificationsData(),
    NotificationService.getNotifications(),
  ])

  return (
    <NotificationsWidget
      initialNotifications={notifications}
      initialUnreadCount={notificationsData.notificationCount}
      totalCount={total}
    />
  )
}
