'use client'

import { Badge } from '@workspace/ui/components/atoms/badge'
import { Button } from '@workspace/ui/components/atoms/button'
import { cn } from '@workspace/ui/lib/utils'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import type { RouteConfig } from '../domain/routes'
import { ICON_MAP, isRouteActive } from '../domain/routes'

export interface NavigationLinkProps {
  route: RouteConfig
}

/**
 * Client component for navigation link with active state detection
 */
export function NavigationLink({ route }: Readonly<NavigationLinkProps>) {
  const pathname = usePathname()
  const isActive = isRouteActive(route.href, pathname ?? '')
  const Icon = ICON_MAP[route.icon]

  return (
    <Button
      asChild
      variant="ghost"
      className={cn(
        'w-full h-9 px-4 justify-start gap-2 mb-1 rounded-lg font-normal',
        isActive
          ? 'bg-sidebar-accent text-sidebar-accent-foreground hover:bg-sidebar-accent'
          : 'text-foreground hover:bg-muted'
      )}
    >
      <Link href={route.href} aria-current={isActive ? 'page' : undefined}>
        <Icon size={14} className="flex-shrink-0" />
        <span className="text-sm font-normal flex-1">{route.label}</span>
        {route.badge && (
          <Badge variant="default" className="text-xs">
            {route.badge}
          </Badge>
        )}
      </Link>
    </Button>
  )
}
