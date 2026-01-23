'use client'

import * as React from 'react'
import { ChecksTag } from '../molecules/checks-tag'
import { NotificationButton } from '../molecules/notification-button'
import { Breadcrumb } from '../molecules/breadcrumb'
import type { BreadcrumbItem } from '../molecules/breadcrumb'
import { cn } from '../../lib/utils'

export interface HeaderProps extends React.HTMLAttributes<HTMLElement> {
  checks?: {
    current: number
    max: number
    refillDate: string
  }
  onNotificationClick?: () => void
  hasNotifications?: boolean
  notificationCount?: number
  breadcrumbs?: BreadcrumbItem[]
}

const Header = React.forwardRef<HTMLElement, HeaderProps>(
  (
    {
      checks,
      onNotificationClick,
      hasNotifications,
      notificationCount,
      breadcrumbs,
      className,
      ...props
    },
    ref
  ) => {
    return (
      <header
        ref={ref}
        className={cn('w-full border-b border-border bg-background', className)}
        {...props}
      >
        <div className="flex items-center justify-between px-24 py-4 gap-4">
          {breadcrumbs && breadcrumbs.length > 0 ? <Breadcrumb items={breadcrumbs} /> : <div />}
          <div className="flex items-center gap-4">
            {checks && (
              <ChecksTag current={checks.current} max={checks.max} refillDate={checks.refillDate} />
            )}
            <NotificationButton
              onClick={onNotificationClick}
              hasNotifications={hasNotifications}
              notificationCount={notificationCount}
            />
          </div>
        </div>
      </header>
    )
  }
)
Header.displayName = 'Header'

export { Header }
