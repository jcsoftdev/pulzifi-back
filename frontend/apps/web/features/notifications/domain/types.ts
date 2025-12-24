/**
 * Notifications Feature - Domain Types
 */

export interface NotificationsData {
  hasNotifications: boolean
  notificationCount: number
}

export interface Notification {
  id: string
  title: string
  message: string
  type: 'info' | 'warning' | 'error' | 'success'
  createdAt: string
  read: boolean
}
