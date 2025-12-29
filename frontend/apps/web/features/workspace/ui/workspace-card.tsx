import { ArrowRight } from 'lucide-react'
import { Card } from '@workspace/ui/components/atoms/card'
import { Button } from '@workspace/ui/components/atoms/button'
import { Badge } from '@workspace/ui/components/atoms/badge'
import { cn } from '@workspace/ui/lib/utils'
import { WorkspaceDropdown } from './workspace-dropdown'
import type { Workspace, WorkspaceType } from '../domain/types'

export interface WorkspaceCardProps {
  workspace: Workspace & {
    pageCount: number
    status: 'Active' | 'Deleted'
  }
  onSelect: (id: string) => void
  onOpen?: (id: string) => void
  onRename?: (id: string) => void
  onEditTag?: (id: string) => void
  onDelete?: (id: string) => void
}

const typeStyles: Record<
  WorkspaceType,
  {
    bg: string
    text: string
  }
> = {
  Personal: {
    bg: 'bg-destructive/10',
    text: 'text-foreground',
  },
  Team: {
    bg: 'bg-accent',
    text: 'text-accent-foreground',
  },
  Competitor: {
    bg: 'bg-accent',
    text: 'text-accent-foreground',
  },
}

const formatCreatedAt = (dateString: string): string => {
  const date = new Date(dateString)
  const now = new Date()
  const diffTime = Math.abs(now.getTime() - date.getTime())
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24))
  const diffMonths = Math.floor(diffDays / 30)
  const diffYears = Math.floor(diffDays / 365)

  if (diffDays < 1) return 'Created today'
  if (diffDays === 1) return 'Created 1 day ago'
  if (diffDays < 30) return `Created ${diffDays} days ago`
  if (diffMonths === 1) return 'Created 1 month ago'
  if (diffMonths < 12) return `Created ${diffMonths} months ago`
  if (diffYears === 1) return 'Created 1 year ago'
  return `Created ${diffYears} years ago`
}

export function WorkspaceCard({
  workspace,
  onSelect,
  onOpen,
  onRename,
  onEditTag,
  onDelete,
}: Readonly<WorkspaceCardProps>) {
  const styles = typeStyles[workspace.type as WorkspaceType]

  return (
    <Card className="w-64 flex flex-col gap-5 shadow-sm rounded-lg py-0">
      <div className="flex flex-col justify-center self-stretch gap-1 p-4 pt-3">
        {/* Header */}
        <div className="flex justify-between items-center self-stretch gap-1">
          <p className="text-xs font-normal text-muted-foreground">
            {formatCreatedAt(workspace.createdAt)}
          </p>
          <WorkspaceDropdown
            trigger={
              <Button
                variant="ghost"
                size="icon-sm"
                className="h-5 w-5 rounded-sm"
                aria-label="Workspace options"
              >
                <svg width="21" height="21" viewBox="0 0 21 21" fill="none" aria-hidden="true">
                  <circle cx="10.5" cy="10.5" r="1" fill="currentColor" />
                  <circle cx="10.5" cy="5.5" r="1" fill="currentColor" />
                  <circle cx="10.5" cy="15.5" r="1" fill="currentColor" />
                </svg>
              </Button>
            }
            onOpen={() => onOpen?.(workspace.id)}
            onRename={() => onRename?.(workspace.id)}
            onEditTag={() => onEditTag?.(workspace.id)}
            onDelete={() => onDelete?.(workspace.id)}
          />
        </div>

        {/* Title and Badge */}
        <div className="flex flex-col gap-1 py-4">
          <div className="flex justify-start items-center gap-1">
            <h3 className="text-xl font-semibold text-foreground">{workspace.name}</h3>
          </div>
          <Badge
            variant="outline"
            className={cn(
              'w-fit px-2 py-1 rounded text-xs font-normal border',
              styles.bg,
              styles.text
            )}
          >
            {workspace.type}
          </Badge>
        </div>
      </div>

      {/* Footer */}
      <div className="flex justify-between items-center self-stretch gap-1 p-3 py-2 border-t border-black/6">
        <div className="flex items-center gap-1">
          <div className="flex justify-center items-center gap-1 p-1 w-6 h-6">
            <svg
              width="13"
              height="13"
              viewBox="0 0 13 13"
              fill="none"
              className="text-foreground"
              aria-hidden="true"
            >
              <rect width="5" height="5" rx="1" fill="currentColor" />
              <rect x="7" width="5" height="5" rx="1" fill="currentColor" />
              <rect y="7" width="5" height="5" rx="1" fill="currentColor" />
            </svg>
          </div>
          <div className="flex flex-col">
            <p className="text-base font-normal text-black/65">0 pages</p>
          </div>
        </div>
        <Button
          variant="ghost"
          size="icon-sm"
          className="h-6 w-6"
          onClick={() => onSelect(workspace.id)}
          aria-label={`View ${workspace.name}`}
        >
          <ArrowRight className="h-4 w-4" />
        </Button>
      </div>
    </Card>
  )
}
