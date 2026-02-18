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
import type { Page } from '../domain/types'

export interface EditPageDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (
    pageId: string,
    data: {
      name: string
      url: string
    }
  ) => Promise<void>
  page: Page | null
  isLoading?: boolean
  error?: Error | null
}

export function EditPageDialog({
  open,
  onOpenChange,
  onSubmit,
  page,
  isLoading = false,
  error,
}: Readonly<EditPageDialogProps>) {
  const [name, setName] = useState('')
  const [url, setUrl] = useState('')
  const nameId = useId()
  const urlId = useId()

  useEffect(() => {
    if (open && page) {
      setName(page.name)
      setUrl(page.url)
    }
  }, [
    open,
    page,
  ])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!page || !name.trim() || !url.trim()) {
      return
    }

    await onSubmit(page.id, {
      name: name.trim(),
      url: url.trim(),
    })
  }

  const isFormValid = name.trim() !== '' && url.trim() !== ''

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Edit Page</DialogTitle>
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
              {isLoading ? 'Saving...' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
