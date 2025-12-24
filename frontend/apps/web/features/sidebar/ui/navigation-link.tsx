'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@workspace/ui'
import type { RouteConfig } from '../domain/routes'
import { isRouteActive } from '../domain/routes'

export interface NavigationLinkProps {
  route: RouteConfig
}

/**
 * Client component for navigation link with active state detection
 */
export function NavigationLink({ route }: Readonly<NavigationLinkProps>) {
  const pathname = usePathname()
  const isActive = isRouteActive(route.href, pathname ?? '')
  const Icon = route.icon

  return (
    <Link
      href={route.href}
      className={cn(
        'w-full px-4 h-[42px] flex items-center gap-2 mb-1 rounded-lg transition-colors',
        isActive
          ? 'bg-sidebar-accent text-sidebar-accent-foreground'
          : 'text-foreground hover:bg-muted'
      )}
      aria-current={isActive ? 'page' : undefined}
    >
      <Icon size={14} className="flex-shrink-0" />
      <span className="text-[14.6px] font-normal flex-1">{route.label}</span>
      {route.badge && (
        <span className="text-[12.5px] px-2 py-0.5 rounded bg-primary text-primary-foreground font-normal">
          {route.badge}
        </span>
      )}
    </Link>
  )
}
