'use client'

import { Pencil, Trash2 } from 'lucide-react'
import { Button } from '@workspace/ui/components/atoms/button'
import type { Page } from '@workspace/services/page-api'

interface PageInfoCardProps {
  page: Page
  onEdit: () => void
  onDelete: () => void
  onViewChanges: () => void
}

export function PageInfoCard({
  page,
  onEdit,
  onDelete,
  onViewChanges,
}: Readonly<PageInfoCardProps>) {
  return (
    <div className="flex flex-col gap-6 bg-card p-6 rounded-xl shadow-sm border border-border">
      <div className="flex flex-col md:flex-row justify-between gap-6">
        <div className="flex gap-4">
          <div className="flex flex-col gap-2">
            <div className="flex items-center gap-2">
              <h1 className="text-2xl font-semibold text-foreground">{page.name}</h1>
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                onClick={onEdit}
                aria-label="Edit page"
                className="text-muted-foreground hover:text-foreground"
              >
                <Pencil className="h-4 w-4" />
              </Button>
            </div>

            <a
              href={page.url}
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm font-normal text-muted-foreground hover:underline flex items-center gap-1"
            >
              {page.url}
            </a>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            className="h-9 px-4 gap-2 bg-transparent border-border text-foreground hover:bg-muted hover:text-foreground"
            onClick={onDelete}
          >
            <Trash2 className="h-4 w-4" />
            Delete
          </Button>
          <Button
            type="button"
            variant="outline"
            onClick={onViewChanges}
            className="px-3 py-2 h-auto gap-2"
            aria-label="View changes"
          >
            <span className="text-sm">View Changes</span>
            <span className="w-2 h-2 rounded-full bg-foreground" />
          </Button>
        </div>
      </div>
    </div>
  )
}
