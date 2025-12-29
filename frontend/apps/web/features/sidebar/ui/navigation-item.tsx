'use client'

import Link from 'next/link'
import type { NavigationItem } from '../domain/types'
import { cn } from '@workspace/ui'

export interface NavigationItemComponentProps {
  item: NavigationItem
}

export function NavigationItemComponent({ item }: Readonly<NavigationItemComponentProps>) {
  return (
    <Link
      href={item.href}
      className={cn(
        'w-full px-4 h-[42px] flex items-center gap-2 mb-1 rounded-lg transition-colors',
        item.active
          ? 'bg-sidebar-accent text-sidebar-primary'
          : 'text-sidebar-foreground hover:bg-sidebar-accent'
      )}
    >
      <div className="flex-shrink-0">{item.icon}</div>
      <span className="text-[14.6px] font-normal flex-1">{item.label}</span>
      {item.badge && (
        <span className="text-[12.5px] px-2 py-0.5 rounded bg-sidebar-primary text-sidebar-primary-foreground font-normal">
          {item.badge}
        </span>
      )}
    </Link>
  )
}
