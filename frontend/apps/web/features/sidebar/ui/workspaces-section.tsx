'use client'

import { useState, useCallback, useEffect, memo } from 'react'
import { usePathname } from 'next/navigation'
import { cn } from '@workspace/ui'
import Link from 'next/link'
import { PlusIcon } from './icons/PlusIcon'
import { MinusIcon } from './icons/MinusIcon'
import { isWorkspaceActive, WORKSPACES_ROUTE } from '../domain/routes'
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
    <Link
      href={`/workspaces/${workspace.id}`}
      className={cn(
        'block px-6 py-2.5 rounded-lg text-[14.6px] transition-colors',
        isActive
          ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium'
          : 'text-foreground hover:bg-muted'
      )}
      aria-current={isActive ? 'page' : undefined}
    >
      {workspace.name}
    </Link>
  )
})

export function WorkspacesSection({ workspaces }: Readonly<WorkspacesSectionClientProps>) {
  const pathname = usePathname()
  const Icon = WORKSPACES_ROUTE.icon

  const isOnWorkspacePath = pathname?.startsWith('/workspaces')
  const [isOpen, setIsOpen] = useState<boolean>(isOnWorkspacePath ?? false)

  useEffect(() => {
    if (isOnWorkspacePath) {
      setIsOpen(true)
    }
  }, [isOnWorkspacePath])

  const toggle = useCallback(() => {
    setIsOpen((prev) => !prev)
  }, [])

  return (
    <div className="mt-2">
      <button
        onClick={toggle}
        className="w-full px-4 py-2.5 rounded-lg flex items-center justify-between text-foreground hover:bg-muted transition-colors"
        aria-expanded={isOpen}
        aria-controls="workspaces-list"
      >
        <div className="flex items-center gap-2">
          <Icon size={14} className="text-foreground" />
          <span className="text-[14.6px] font-normal">{WORKSPACES_ROUTE.label}</span>
        </div>

        <span className="inline-flex items-center justify-center w-6 h-6 relative">
          <MinusIcon
            className={cn(
              'absolute inset-0 m-auto text-primary transition-opacity duration-200',
              isOpen ? 'opacity-100' : 'opacity-0'
            )}
          />
          <PlusIcon
            className={cn(
              'absolute inset-0 m-auto text-primary transition-opacity duration-200',
              isOpen ? 'opacity-0' : 'opacity-100'
            )}
          />
        </span>
      </button>

      <div
        id="workspaces-list"
        className={cn(
          'overflow-hidden mt-1 bg-muted/50 rounded-lg space-y-1',
          'transition-[max-height,opacity,padding] duration-300 ease-in-out',
          isOpen
            ? 'max-h-96 opacity-100 py-1 px-6'
            : 'max-h-0 opacity-0 py-0 px-6'
        )}
      >
        {workspaces.map((workspace) => {
          const isActive = isWorkspaceActive(workspace.id, pathname ?? '')
          return (
            <WorkspaceItem
              key={workspace.id}
              workspace={workspace}
              isActive={isActive}
            />
          )
        })}
      </div>
    </div>
  )
}
