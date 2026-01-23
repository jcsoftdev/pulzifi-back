'use client'

import { useEffect } from 'react'
import type { BreadcrumbItem } from '@workspace/ui/components/molecules'

interface WorkspaceBreadcrumbsProps {
  breadcrumbs: BreadcrumbItem[]
}

export function WorkspaceBreadcrumbs({ breadcrumbs }: Readonly<WorkspaceBreadcrumbsProps>) {
  useEffect(() => {
    // Update breadcrumbs in the app shell
    const event = new CustomEvent('updateBreadcrumbs', {
      detail: {
        breadcrumbs,
      },
    })
    window.dispatchEvent(event)

    // Cleanup: reset breadcrumbs when unmounting
    return () => {
      const resetEvent = new CustomEvent('updateBreadcrumbs', {
        detail: {
          breadcrumbs: [],
        },
      })
      window.dispatchEvent(resetEvent)
    }
  }, [
    breadcrumbs,
  ])

  return null
}
