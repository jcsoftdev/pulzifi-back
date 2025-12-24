'use client'

import { ChevronDown } from 'lucide-react'
import type { Organization } from '../domain/types'

export interface OrganizationSelectorProps {
  organization: Organization
  onSelect?: () => void
}

export function OrganizationSelector({ organization, onSelect }: Readonly<OrganizationSelectorProps>) {
  return (
    <div className="px-3 py-1">
      <button
        onClick={onSelect}
        className="w-full p-2 border border-border rounded-lg flex items-center gap-2 hover:bg-muted transition-colors bg-card"
      >
        <div className="w-[33px] h-[33px] rounded-full border border-border bg-muted flex items-center justify-center text-muted-foreground font-normal text-xs flex-shrink-0 overflow-hidden">
          {organization.name.charAt(0).toUpperCase()}
        </div>
        <div className="flex-1 text-left min-w-0">
          <p className="text-[12.5px] font-normal text-foreground truncate leading-tight">{organization.name}</p>
          {organization.company && (
            <p className="text-[14.6px] font-medium text-foreground truncate leading-tight">{organization.company}</p>
          )}
        </div>
        <ChevronDown className="w-[12.5px] h-[12.5px] text-foreground flex-shrink-0" />
      </button>
    </div>
  )
}
