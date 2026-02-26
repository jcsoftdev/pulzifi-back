'use client'

import { X } from 'lucide-react'
import * as React from 'react'

import { cn } from '../../lib/utils'
import { Badge } from './badge'

interface TagsInputProps {
  value: string[]
  onChange: (tags: string[]) => void
  disabled?: boolean
  placeholder?: string
  className?: string
}

export function TagsInput({
  value,
  onChange,
  disabled,
  placeholder = 'Add tag…',
  className,
}: Readonly<TagsInputProps>) {
  const [inputValue, setInputValue] = React.useState('')
  const inputRef = React.useRef<HTMLInputElement>(null)

  const addTag = (raw: string) => {
    const trimmed = raw.trim()
    if (trimmed && !value.includes(trimmed)) {
      onChange([...value, trimmed])
    }
    setInputValue('')
  }

  const removeTag = (tag: string) => {
    onChange(value.filter((t) => t !== tag))
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault()
      addTag(inputValue)
    } else if (e.key === 'Backspace' && !inputValue && value.length > 0) {
      const lastTag = value.at(-1)
      if (lastTag) {
        removeTag(lastTag)
      }
    }
  }

  return (
    <fieldset
      className={cn(
        'flex flex-wrap gap-1.5 min-h-9 w-full rounded-md border border-input bg-background px-3 py-2 ring-offset-background focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-2 cursor-text',
        disabled && 'cursor-not-allowed opacity-50',
        className
      )}
      disabled={disabled}
    >
      {value.map((tag) => (
        <Badge key={tag} variant="secondary" className="gap-1 h-5 px-2 text-xs font-normal">
          {tag}
          {!disabled && (
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation()
                removeTag(tag)
              }}
              className="ml-0.5 rounded-full outline-none focus-visible:ring-1 focus-visible:ring-ring"
            >
              <X className="h-3 w-3" />
              <span className="sr-only">Remove {tag}</span>
            </button>
          )}
        </Badge>
      ))}
      <input
        ref={inputRef}
        value={inputValue}
        onChange={(e) => setInputValue(e.target.value)}
        onKeyDown={handleKeyDown}
        onBlur={() => {
          if (inputValue.trim()) addTag(inputValue)
        }}
        placeholder={value.length === 0 ? placeholder : ''}
        disabled={disabled}
        className="flex-1 min-w-[80px] bg-transparent text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed"
      />
    </fieldset>
  )
}
