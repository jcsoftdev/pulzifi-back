/**
 * Notification Service - Domain Service
 * Handles all notification-related data fetching
 * This is feature-specific and belongs to the Notifications domain
 */

import { NotificationApi } from '@workspace/services'
import { UnauthorizedError } from '@workspace/shared-http'
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
      if (error instanceof UnauthorizedError) {
        return handleServerAuthError(error)
      }
      console.error('Failed to fetch notification count:', error)
      return { hasNotifications: false, notificationCount: 0 }
    }
  },

  /**
   * Get all notifications for the current user
   */
  async getNotifications(): Promise<{
    notifications: Notification[]
    total: number
  }> {
    try {
      return await NotificationApi.getNotifications()
    } catch (error) {
      if (error instanceof UnauthorizedError) {
        return handleServerAuthError(error)
      }
      console.error('Failed to fetch notifications:', error)
      return { notifications: [], total: 0 }
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

  /**
   * Mark all notifications as read
   */
  async markAllAsRead(): Promise<void> {
    try {
      await NotificationApi.markAllAsRead()
    } catch (error) {
      console.error('Failed to mark all notifications as read:', error)
      throw error
    }
  },
}
