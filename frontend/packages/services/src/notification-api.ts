/**
 * Notification API Service
 * Consumes backend /api/v1/alerts/* endpoints
 * Works dynamically in both server-side and client-side
 */

import { getHttpClient } from '@workspace/shared-http'

export interface NotificationsData {
  hasNotifications: boolean
  notificationCount: number
}

export interface Notification {
  id: string
  workspaceId: string
  pageId: string
  checkId: string
  title: string
  message: string
  type: 'info' | 'warning' | 'error' | 'success'
  read: boolean
  createdAt: string
  pageName?: string
  pageUrl?: string
}

interface UnreadCountDTO {
  has_notifications: boolean
  notification_count: number
}

interface AlertItemDTO {
  id: string
  workspace_id: string
  page_id: string
  check_id: string
  title: string
  description: string
  type: string
  read: boolean
  created_at: string
  page_name: string
  page_url: string
}

interface ListAllAlertsDTO {
  data: AlertItemDTO[]
  total: number
}

function mapAlertType(type_: string): Notification['type'] {
  switch (type_) {
    case 'content_change':
      return 'warning'
    case 'error':
      return 'error'
    default:
      return 'info'
  }
}

export const NotificationApi = {
  async getNotificationsData(): Promise<NotificationsData> {
    const http = await getHttpClient()
    const dto = await http.get<UnreadCountDTO>('/api/v1/alerts/unread-count')
    return {
      hasNotifications: dto.has_notifications,
      notificationCount: dto.notification_count,
    }
  },

  async getNotifications(): Promise<{
    notifications: Notification[]
    total: number
  }> {
    const http = await getHttpClient()
    const dto = await http.get<ListAllAlertsDTO>('/api/v1/alerts/all')
    return {
      notifications: dto.data.map((item) => ({
        id: item.id,
        workspaceId: item.workspace_id,
        pageId: item.page_id,
        checkId: item.check_id,
        title: item.title,
        message: item.description,
        type: mapAlertType(item.type),
        read: item.read,
        createdAt: item.created_at,
        pageName: item.page_name,
        pageUrl: item.page_url,
      })),
      total: dto.total,
    }
  },

  async markAsRead(notificationId: string): Promise<void> {
    const http = await getHttpClient()
    await http.put(`/api/v1/alerts/${notificationId}`)
  },

  async markAllAsRead(): Promise<void> {
    const http = await getHttpClient()
    await http.put('/api/v1/alerts/read-all')
  },
}
