'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { cn } from '@workspace/ui/lib/utils'
import { ChevronDown, SquarePlus } from 'lucide-react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { memo, useCallback, useEffect, useState } from 'react'
import { ICON_MAP, isWorkspaceActive, WORKSPACES_ROUTE } from '../domain/routes'
import type { Workspace } from '../domain/types'

export interface WorkspacesSectionClientProps {
  workspaces: Workspace[]
}

const WorkspaceItem = memo(function WorkspaceItem({
  workspace,
  isActive,
}: {
  workspace: Workspace
  isActive: boolean
}) {
  return (
    <Button
      asChild
      variant="ghost"
      className={cn(
        'w-full h-auto px-6 py-2.5 rounded-lg text-sm justify-start font-normal',
        isActive
          ? 'bg-sidebar-accent text-sidebar-accent-foreground hover:bg-sidebar-accent font-medium'
          : 'text-foreground hover:bg-muted'
      )}
    >
      <Link href={`/workspaces/${workspace.id}`} aria-current={isActive ? 'page' : undefined}>
        {workspace.name}
      </Link>
    </Button>
  )
})

export function WorkspacesSection({ workspaces }: Readonly<WorkspacesSectionClientProps>) {
  const pathname = usePathname()
  const Icon = ICON_MAP[WORKSPACES_ROUTE.icon]

  const isOnWorkspacePath = pathname?.startsWith('/workspaces')
  const [isOpen, setIsOpen] = useState<boolean>(isOnWorkspacePath ?? false)

  const listId = 'workspaces-list'

  useEffect(() => {
    if (isOnWorkspacePath) {
      setIsOpen(true)
    }
  }, [
    isOnWorkspacePath,
  ])

  const toggle = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsOpen((prev) => !prev)
  }, [])

  const handleCreateWorkspace = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    console.log('Create workspace clicked')
  }, [])

  return (
    <div className="mt-2">
      <div
        className={cn(
          'w-full px-4 py-2.5 rounded-lg flex items-center justify-between transition-colors',
          isOnWorkspacePath
            ? 'bg-sidebar-accent text-sidebar-accent-foreground'
            : 'text-foreground hover:bg-muted'
        )}
      >
        <Link href="/workspaces" className="flex items-center gap-2 flex-1">
          <Icon
            size={14}
            className={isOnWorkspacePath ? 'text-sidebar-accent-foreground' : 'text-foreground'}
          />
          <span className={cn('text-sm', isOnWorkspacePath ? 'font-semibold' : 'font-normal')}>
            {WORKSPACES_ROUTE.label}
          </span>
        </Link>

        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={handleCreateWorkspace}
            className="h-6 w-6 hover:bg-sidebar-primary/10"
            aria-label="Create workspace"
          >
            <SquarePlus
              size={16}
              className={
                isOnWorkspacePath ? 'text-sidebar-accent-foreground' : 'text-sidebar-primary'
              }
            />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={toggle}
            className="h-6 w-6 hover:bg-sidebar-primary/10"
            aria-expanded={isOpen}
            aria-controls={listId}
            aria-label={isOpen ? 'Collapse workspaces' : 'Expand workspaces'}
          >
            <ChevronDown
              size={16}
              className={cn(
                'transition-transform duration-200',
                isOpen ? 'rotate-180 text-sidebar-accent-foreground' : 'text-foreground'
              )}
            />
          </Button>
        </div>
      </div>

      <div
        id={listId}
        className={cn(
          'overflow-hidden mt-1 bg-muted/50 rounded-lg space-y-1',
          'transition-[max-height,opacity,padding] duration-300 ease-in-out',
          isOpen ? 'max-h-96 opacity-100 py-1 px-6' : 'max-h-0 opacity-0 py-0 px-6'
        )}
      >
        {workspaces.map((workspace) => {
          const isActive = isWorkspaceActive(workspace.id, pathname ?? '')
          return <WorkspaceItem key={workspace.id} workspace={workspace} isActive={isActive} />
        })}
      </div>
    </div>
  )
}
