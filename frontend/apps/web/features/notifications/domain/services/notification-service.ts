/**
 * Notification Service - Domain Service
 * Handles all notification-related data fetching
 * This is feature-specific and belongs to the Notifications domain
 */

import { NotificationApi } from '@workspace/services'
import type { NotificationsData, Notification } from '../types'

export class NotificationService {
  /**
   * Get notification count and status
   * Works in both server-side and client-side
   */
  static async getNotificationsData(): Promise<NotificationsData> {
    try {
      const data = await NotificationApi.getNotificationsData()
      return data
    } catch (error) {
      console.error('Failed to fetch notifications data:', error)
      
      // Fallback to mock data
      return {
        hasNotifications: true,
        notificationCount: 3
      }
    }
  }

  /**
   * Get all notifications for the current user
   */
  static async getNotifications(): Promise<Notification[]> {
    try {
      const data = await NotificationApi.getNotifications()
      return data
    } catch (error) {
      console.error('Failed to fetch notifications:', error)
      return []
    }
  }

  /**
   * Mark notification as read
   */
  static async markAsRead(notificationId: string): Promise<void> {
    try {
      await NotificationApi.markAsRead(notificationId)
    } catch (error) {
      console.error('Failed to mark notification as read:', error)
      throw error
    }
  }
}
