'use client'

import { ChevronDown } from 'lucide-react'
import { Button } from '@workspace/ui/components/atoms/button'
import { Avatar, AvatarFallback } from '@workspace/ui/components/atoms/avatar'
import type { Organization } from '../domain/types'

export interface OrganizationSelectorProps {
  organization: Organization
  onSelect?: () => void
}

export function OrganizationSelector({
  organization,
  onSelect,
}: Readonly<OrganizationSelectorProps>) {
  return (
    <div className="px-3 py-1">
      <Button
        variant="outline"
        onClick={onSelect}
        className="w-full h-auto p-2 justify-start gap-2 bg-card font-normal"
      >
        <Avatar className="w-8 h-8 rounded-full">
          <AvatarFallback className="text-xs">
            {organization.name.charAt(0).toUpperCase()}
          </AvatarFallback>
        </Avatar>
        <div className="flex-1 text-left min-w-0">
          <p className="text-xs font-normal text-foreground truncate leading-tight">
            {organization.name}
          </p>
          {organization.company && (
            <p className="text-sm font-medium text-foreground truncate leading-tight">
              {organization.company}
            </p>
          )}
        </div>
        <ChevronDown className="w-3 h-3 text-foreground flex-shrink-0" />
      </Button>
    </div>
  )
}
