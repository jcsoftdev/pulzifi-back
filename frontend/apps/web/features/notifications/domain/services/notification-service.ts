/**
 * Notification Service - Domain Service
 * Handles all notification-related data fetching
 * This is feature-specific and belongs to the Notifications domain
 */

import { NotificationApi } from '@workspace/services'
import type { Notification, NotificationsData } from '../types'

export const NotificationService = {
  /**
   * Get notification count and status
   * Works in both server-side and client-side
   */
  async getNotificationsData(): Promise<NotificationsData> {
    return await NotificationApi.getNotificationsData()
  },

  /**
   * Get all notifications for the current user
   */
  async getNotifications(): Promise<Notification[]> {
    return await NotificationApi.getNotifications()
  },

  /**
   * Mark notification as read
   */
  async markAsRead(notificationId: string): Promise<void> {
    try {
      await NotificationApi.markAsRead(notificationId)
    } catch (error) {
      console.error('Failed to mark notification as read:', error)
      throw error
    }
  },
}
