'use client'

import { Avatar, AvatarFallback, AvatarImage } from '@workspace/ui/components/atoms/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@workspace/ui/components/molecules/dropdown-menu'
import { MoreHorizontal, Pencil, Trash2 } from 'lucide-react'
import { useState } from 'react'
import type { TeamMember } from '../domain/types'
import { memberFullName } from '../domain/types'

interface MemberCardProps {
  member: TeamMember
  isCurrentUser: boolean
  onEdit: (member: TeamMember) => void
  onDelete: (member: TeamMember) => void
}

function getInitials(firstName: string, lastName: string): string {
  return `${firstName.charAt(0)}${lastName.charAt(0)}`.toUpperCase()
}

export function MemberCard({ member, isCurrentUser, onEdit, onDelete }: Readonly<MemberCardProps>) {
  const [menuOpen, setMenuOpen] = useState(false)

  const displayName = memberFullName(member)
  const label = isCurrentUser ? `${displayName} (You)` : displayName

  return (
    <div className="group relative flex flex-col items-center gap-2 w-20">
      {/* Avatar */}
      <div className="relative">
        <Avatar className="w-16 h-16 ring-2 ring-border">
          {member.avatarUrl && <AvatarImage src={member.avatarUrl} alt={displayName} />}
          <AvatarFallback className="text-sm font-medium bg-muted text-muted-foreground">
            {getInitials(member.firstName, member.lastName)}
          </AvatarFallback>
        </Avatar>

        {/* Actions menu - shown on hover or when open, hidden for current user (owner) */}
        {!isCurrentUser && member.role !== 'OWNER' && (
          <DropdownMenu open={menuOpen} onOpenChange={setMenuOpen}>
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                className="absolute -top-1 -right-1 p-0.5 rounded-full bg-background border border-border text-muted-foreground opacity-0 group-hover:opacity-100 focus:opacity-100 transition-opacity"
                aria-label={`Actions for ${displayName}`}
              >
                <MoreHorizontal className="w-3.5 h-3.5" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-36">
              <DropdownMenuItem onClick={() => onEdit(member)}>
                <Pencil className="w-4 h-4" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem variant="destructive" onClick={() => onDelete(member)}>
                <Trash2 className="w-4 h-4" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </div>

      {/* Name */}
      <span className="text-xs text-center text-foreground leading-tight max-w-full truncate" title={label}>
        {label}
      </span>
    </div>
  )
}
