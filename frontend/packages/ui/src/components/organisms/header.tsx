'use client'

import * as React from "react"
import { ChecksTag } from "../molecules/checks-tag"
import { NotificationButton } from "../molecules/notification-button"
import { cn } from "../../lib/utils"

export interface HeaderProps extends React.HTMLAttributes<HTMLElement> {
  checks?: {
    current: number
    max: number
    refillDate: string
  }
  onNotificationClick?: () => void
  hasNotifications?: boolean
  notificationCount?: number
}

const Header = React.forwardRef<HTMLElement, HeaderProps>(
  ({ checks, onNotificationClick, hasNotifications, notificationCount, className, ...props }, ref) => {
    return (
      <header
        ref={ref}
        className={cn("w-full border-b border-border bg-background", className)}
        {...props}
      >
        <div className="flex items-center justify-end px-24 py-4 gap-4">
          {checks && (
            <ChecksTag
              current={checks.current}
              max={checks.max}
              refillDate={checks.refillDate}
            />
          )}
          <NotificationButton
            onClick={onNotificationClick}
            hasNotifications={hasNotifications}
            notificationCount={notificationCount}
          />
        </div>
      </header>
    )
  }
)
Header.displayName = "Header"

export { Header }
