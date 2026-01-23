'use client'

import { Pencil, Trash2, ArrowUpRight } from 'lucide-react'
import { Button } from '@workspace/ui/components/atoms/button'
import { Badge } from '@workspace/ui/components/atoms/badge'
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
              <button
                onClick={onEdit}
                className="p-1 text-muted-foreground hover:text-foreground transition-colors"
              >
                <Pencil className="h-4 w-4" />
              </button>
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
            className="h-10 px-4 gap-2 bg-transparent border-border text-foreground hover:bg-muted hover:text-foreground"
            onClick={onDelete}
          >
            <Trash2 className="h-4 w-4" />
            Delete
          </Button>
          <button
            onClick={onViewChanges}
            className="flex items-center gap-2 px-3 py-2 rounded-md border border-border hover:bg-muted transition-colors"
            aria-label="View changes"
          >
            <span className="text-sm text-foreground">View Changes</span>
            <div className="w-2 h-2 rounded-full bg-foreground" />
          </button>
        </div>
      </div>
    </div>
  )
}
