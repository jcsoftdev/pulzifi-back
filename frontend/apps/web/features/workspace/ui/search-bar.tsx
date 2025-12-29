'use client'

import { Search } from 'lucide-react'

export interface SearchBarProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
}

export function SearchBar({
  value,
  onChange,
  placeholder = 'Search workspaces',
}: Readonly<SearchBarProps>) {
  return (
    <div className="flex justify-stretch items-stretch gap-2.5 px-3 w-96 h-8 rounded-md border border-border bg-background">
      <div className="flex items-center gap-2 py-1 flex-1 h-8">
        <Search className="w-4 h-4 text-foreground" />
        <input
          type="text"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          className="flex-1 text-sm font-normal text-foreground placeholder:text-muted-foreground/50 bg-transparent border-none outline-none"
        />
      </div>
    </div>
  )
}
