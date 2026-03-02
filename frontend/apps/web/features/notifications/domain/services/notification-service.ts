/**
 * Notification Service - Domain Service
 * Handles all notification-related data fetching
 * This is feature-specific and belongs to the Notifications domain
 */

import { NotificationApi } from '@workspace/services'
import { handleServerAuthError } from '@/lib/auth/server-auth'
import type { Notification, NotificationsData } from '../types'

export const NotificationService = {
  /**
   * Get notification count and status
   * Works in both server-side and client-side
   */
  async getNotificationsData(): Promise<NotificationsData> {
    try {
      return await NotificationApi.getNotificationsData()
    } catch (error) {
      return handleServerAuthError(error)
    }
  },

  /**
   * Get all notifications for the current user
   */
  async getNotifications(): Promise<Notification[]> {
    try {
      return await NotificationApi.getNotifications()
    } catch (error) {
      return handleServerAuthError(error)
    }
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
