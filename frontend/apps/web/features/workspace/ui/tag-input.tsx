'use client'

import { useState } from 'react'
import { X, Plus } from 'lucide-react'
import { Badge } from '@workspace/ui/components/atoms/badge'
import { Input } from '@workspace/ui/components/atoms/input'
import { Button } from '@workspace/ui/components/atoms/button'

interface TagInputProps {
  tags: string[]
  onChange: (tags: string[]) => void
  disabled?: boolean
}

export function TagInput({ tags, onChange, disabled }: TagInputProps) {
  const [inputValue, setInputValue] = useState('')

  const handleAddTag = () => {
    const trimmed = inputValue.trim()
    if (trimmed && !tags.includes(trimmed)) {
      onChange([
        ...tags,
        trimmed,
      ])
      setInputValue('')
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleAddTag()
    }
  }

  const handleRemoveTag = (tagToRemove: string) => {
    onChange(tags.filter((tag) => tag !== tagToRemove))
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap gap-2 mb-2">
        {tags.map((tag) => (
          <Badge key={tag} variant="secondary" className="gap-1 pr-1">
            {tag}
            <button
              type="button"
              onClick={() => handleRemoveTag(tag)}
              disabled={disabled}
              className="ml-1 rounded-full hover:bg-muted p-0.5"
            >
              <X className="h-3 w-3" />
              <span className="sr-only">Remove {tag} tag</span>
            </button>
          </Badge>
        ))}
      </div>
      <div className="flex gap-2">
        <Input
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Add a tag..."
          disabled={disabled}
          className="flex-1"
        />
        <Button
          type="button"
          variant="outline"
          size="icon"
          onClick={handleAddTag}
          disabled={disabled || !inputValue.trim()}
        >
          <Plus className="h-4 w-4" />
          <span className="sr-only">Add tag</span>
        </Button>
      </div>
    </div>
  )
}
