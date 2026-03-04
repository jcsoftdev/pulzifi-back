'use client'

import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { NotificationApi } from '@workspace/services'
import { NotificationButton } from '@workspace/ui/components/molecules'
import { NotixAnchor } from '@workspace/notix'
import type { Notification } from '../domain/types'

function getNotificationHref(notification: Notification): string {
  return `/workspaces/${notification.workspaceId}/pages/${notification.pageId}/changes?checkId=${notification.checkId}`
}

function timeAgo(dateStr: string): string {
  const seconds = Math.floor(
    (Date.now() - new Date(dateStr).getTime()) / 1000
  )
  if (seconds < 60) return 'just now'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

interface NotificationsWidgetProps {
  initialNotifications: Notification[]
  initialUnreadCount: number
  totalCount: number
}

export function NotificationsWidget({
  initialNotifications,
  initialUnreadCount,
  totalCount,
}: NotificationsWidgetProps) {
  const [notifications, setNotifications] = useState(initialNotifications)
  const [unreadCount, setUnreadCount] = useState(initialUnreadCount)
  const router = useRouter()

  const handleClick = (e: React.MouseEvent, notification: Notification) => {
    e.stopPropagation()

    if (!notification.read) {
      setNotifications((prev) =>
        prev.map((n) => (n.id === notification.id ? { ...n, read: true } : n))
      )
      setUnreadCount((c) => Math.max(0, c - 1))

      NotificationApi.markAsRead(notification.id).catch(() => {
        setNotifications((prev) =>
          prev.map((n) =>
            n.id === notification.id ? { ...n, read: false } : n
          )
        )
        setUnreadCount((c) => c + 1)
      })
    }

    router.push(getNotificationHref(notification))
  }

  const remaining = totalCount - notifications.length

  const description =
    notifications.length === 0 ? (
      <p className="text-sm text-muted-foreground py-2">No notifications.</p>
    ) : (
      <div className="flex flex-col divide-y divide-border max-h-64 overflow-y-auto -mx-3">
        {notifications.map((n) => (
          <button
            type="button"
            key={n.id}
            onClick={(e) => handleClick(e, n)}
            className="flex flex-col gap-0.5 px-3 py-2 hover:bg-muted/50 transition-colors text-left"
          >
            <div className="flex items-center justify-between gap-2">
              <span className="text-sm font-medium text-foreground truncate">
                {n.title}
              </span>
              {!n.read && (
                <span className="size-2 shrink-0 rounded-full bg-blue-500" />
              )}
            </div>
            <span className="text-xs text-muted-foreground truncate">
              {n.pageName || n.message}
            </span>
            <span className="text-xs text-muted-foreground/60">
              {timeAgo(n.createdAt)}
            </span>
          </button>
        ))}
        {remaining > 0 && (
          <div className="px-3 py-2 text-xs text-muted-foreground text-center">
            +{remaining} notifications
          </div>
        )}
      </div>
    )

  return (
    <NotixAnchor
      as={NotificationButton}
      hasNotifications={unreadCount > 0}
      notificationCount={unreadCount}
      title={`${unreadCount} Notification${unreadCount === 1 ? '' : 's'}`}
      description={description}
      state="info"
      classNames={{
        title: 'text-sm font-semibold text-foreground',
        description: 'text-sm text-muted-foreground',
      }}
    />
  )
}
