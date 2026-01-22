'use client'

import * as React from 'react'
import { ChevronRight } from 'lucide-react'
import Link from 'next/link'
import { cn } from '../../lib/utils'

export interface BreadcrumbItem {
  label: string
  href?: string
  icon?: React.ReactNode
  isCurrent?: boolean
}

export interface BreadcrumbProps {
  items: BreadcrumbItem[]
  className?: string
}

export function Breadcrumb({ items, className }: Readonly<BreadcrumbProps>) {
  return (
    <nav aria-label="Breadcrumb" className={cn('flex items-center gap-1', className)}>
      <ol className="flex items-center flex-wrap gap-1">
        {items.map((item, index) => {
          const isCurrent = item.isCurrent || index === items.length - 1
          const isFirst = index === 0

          return (
            <React.Fragment key={`${item.label}-${index}`}>
              {index > 0 && (
                <li aria-hidden="true">
                  <ChevronRight className="w-4 h-4 text-foreground/45" />
                </li>
              )}
              <li>
                {isCurrent ? (
                  <span
                    className={cn(
                      'inline-flex items-center gap-1 px-1 py-0.5 rounded text-sm font-normal',
                      'text-foreground/88'
                    )}
                    aria-current="page"
                  >
                    {item.icon && <span className="w-4 h-4">{item.icon}</span>}
                    <span>{item.label}</span>
                  </span>
                ) : (
                  <Link
                    href={item.href || '#'}
                    className={cn(
                      'inline-flex items-center gap-1 px-1 py-0.5 rounded text-sm font-normal',
                      'text-foreground/45 hover:text-foreground/88 hover:bg-muted transition-colors'
                    )}
                  >
                    {item.icon && <span className="w-4 h-4">{item.icon}</span>}
                    <span>{item.label}</span>
                  </Link>
                )}
              </li>
            </React.Fragment>
          )
        })}
      </ol>
    </nav>
  )
}

// Export default breadcrumb icon
export { Home as BreadcrumbHomeIcon } from 'lucide-react'
