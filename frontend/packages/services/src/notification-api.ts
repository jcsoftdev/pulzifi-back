/**
 * Notification API Service
 * Consumes backend /api/notifications/* endpoints
 * Works dynamically in both server-side and client-side
 */

import { getHttpClient } from '@workspace/shared-http'

export interface NotificationsData {
  hasNotifications: boolean
  notificationCount: number
}

export interface Notification {
  id: string
  title: string
  message: string
  type: 'info' | 'warning' | 'error' | 'success'
  read: boolean
  createdAt: string
}

export const NotificationApi = {
  async getNotificationsData(): Promise<NotificationsData> {
    // Mocked for development: backend not ready / or prefer mocked data for now.
    // TODO: replace with real API call when backend endpoint is available.
    return {
      hasNotifications: true,
      notificationCount: 3,
    }
    // Uncomment to call real backend:
    // const http = await getHttpClient()
    // return http.get<NotificationsData>('/api/notifications/count')
  },

  async getNotifications(): Promise<Notification[]> {
    const http = await getHttpClient()
    return http.get<Notification[]>('/api/notifications')
  },

  async markAsRead(notificationId: string): Promise<void> {
    const http = await getHttpClient()
    await http.put(`/api/notifications/${notificationId}/read`)
  },

  async markAllAsRead(): Promise<void> {
    const http = await getHttpClient()
    await http.put('/api/notifications/read-all')
  },
}
