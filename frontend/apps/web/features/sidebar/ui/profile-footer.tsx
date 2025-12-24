'use client'

import type { User } from '../domain/types'

export interface ProfileFooterProps {
  user: User
  onLogout?: () => void
  onSettings?: () => void
}

export function ProfileFooter({ user, onLogout, onSettings }: Readonly<ProfileFooterProps>) {
  return (
    <div className="p-2">
      <div className="flex items-center gap-2 p-2 border border-border rounded-lg bg-card">
        <div className="w-[33px] h-[33px] rounded-lg border border-border bg-muted flex items-center justify-center flex-shrink-0 overflow-hidden">
          {user.avatar ? (
            <img src={user.avatar} alt={user.name} className="w-full h-full object-cover" />
          ) : (
            <span className="text-xs font-normal text-muted-foreground">{user.name.charAt(0).toUpperCase()}</span>
          )}
        </div>
        <div className="flex-1 min-w-0">
          <p className="text-[14.6px] font-semibold text-foreground truncate leading-tight">{user.name}</p>
          <p className="text-[12px] font-normal text-muted-foreground truncate leading-tight">{user.role}</p>
        </div>
        <button className="w-4 h-4 text-foreground flex-shrink-0">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
            <circle cx="8" cy="2.5" r="1" fill="currentColor" />
            <circle cx="8" cy="8" r="1" fill="currentColor" />
            <circle cx="8" cy="13.5" r="1" fill="currentColor" />
          </svg>
        </button>
      </div>
    </div>
  )
}
