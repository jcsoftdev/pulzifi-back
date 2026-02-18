'use client'

import { Bell } from 'lucide-react'
import * as React from 'react'
import { cn } from '../../lib/utils'
import { Button } from '../atoms/button'

export interface NotificationButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  hasNotifications?: boolean
  notificationCount?: number
}

const NotificationButton = React.forwardRef<HTMLButtonElement, NotificationButtonProps>(
  ({ hasNotifications = false, notificationCount, className, ...props }, ref) => {
    return (
      <Button
        ref={ref}
        className={cn(
          'relative w-8 h-8 rounded-md flex items-center justify-center bg-background hover:bg-muted border border-border transition-colors',
          className
        )}
        {...props}
      >
        <Bell className="h-5 w-5 text-foreground" />
        {hasNotifications && (
          <span className="absolute top-1 right-1 flex h-2 w-2">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75" />
            <span className="relative inline-flex rounded-full h-2 w-2 bg-red-500" />
          </span>
        )}
      </Button>
    )
  }
)
NotificationButton.displayName = 'NotificationButton'

export { NotificationButton }
