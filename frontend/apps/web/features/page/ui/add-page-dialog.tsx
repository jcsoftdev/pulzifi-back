'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@workspace/ui/components/atoms/dialog'
import { Input } from '@workspace/ui/components/atoms/input'
import { Label } from '@workspace/ui/components/atoms/label'
import { useEffect, useId, useState } from 'react'
import type { CreatePageDto } from '../domain/types'

export interface AddPageDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: CreatePageDto) => Promise<void>
  workspaceId: string
  isLoading?: boolean
  error?: Error | null
}

export function AddPageDialog({
  open,
  onOpenChange,
  onSubmit,
  workspaceId,
  isLoading = false,
  error,
}: Readonly<AddPageDialogProps>) {
  const [name, setName] = useState('')
  const [url, setUrl] = useState('')
  const nameId = useId()
  const urlId = useId()

  useEffect(() => {
    if (!open) {
      // Reset form when dialog closes
      setName('')
      setUrl('')
    }
  }, [
    open,
  ])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!name.trim() || !url.trim()) {
      return
    }

    const data: CreatePageDto = {
      workspaceId,
      name: name.trim(),
      url: url.trim(),
    }

    await onSubmit(data)
  }

  const isFormValid = name.trim() !== '' && url.trim() !== ''

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Add New Page</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="space-y-4 py-4">
            {/* Page Name */}
            <div className="space-y-2">
              <Label htmlFor={nameId}>Page Name *</Label>
              <Input
                id={nameId}
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g., Home page"
                disabled={isLoading}
                required
              />
            </div>

            {/* Page URL */}
            <div className="space-y-2">
              <Label htmlFor={urlId}>Page URL *</Label>
              <Input
                id={urlId}
                type="url"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder="https://example.com/page"
                disabled={isLoading}
                required
              />
            </div>

            {error && <div className="text-sm text-destructive">{error.message}</div>}
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading || !isFormValid}>
              {isLoading ? 'Adding...' : 'Add Page'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
