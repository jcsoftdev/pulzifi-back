'use client'

import { useState, useEffect, useId } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@workspace/ui/components/atoms/dialog'
import { Button } from '@workspace/ui/components/atoms/button'
import { Input } from '@workspace/ui/components/atoms/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@workspace/ui/components/atoms/select'
import { TagInput } from './tag-input'
import type { Workspace, WorkspaceType } from '@/features/workspace/domain/types'

export interface EditWorkspaceDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (
    id: string,
    data: {
      name: string
      type: WorkspaceType
      tags: string[]
    }
  ) => Promise<void>
  isLoading: boolean
  error: Error | null
  workspace: Workspace
}

export function EditWorkspaceDialog({
  open,
  onOpenChange,
  onSubmit,
  isLoading,
  error,
  workspace,
}: Readonly<EditWorkspaceDialogProps>) {
  const nameId = useId()
  const typeId = useId()
  const [name, setName] = useState(workspace.name || '')
  const [type, setType] = useState<WorkspaceType>((workspace.type as WorkspaceType) || 'Personal')
  const [tags, setTags] = useState<string[]>(workspace.tags || [])
  const [isSubmitting, setIsSubmitting] = useState(false)

  // Reset form when workspace changes or dialog opens
  useEffect(() => {
    if (open) {
      setName(workspace.name || '')
      setType((workspace.type as WorkspaceType) || 'Personal')
      setTags(workspace.tags || [])
    }
  }, [
    open,
    workspace,
  ])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name?.trim()) return

    setIsSubmitting(true)
    try {
      await onSubmit(workspace.id, {
        name: name.trim(),
        type,
        tags,
      })
      onOpenChange(false)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Edit Workspace</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <label htmlFor={nameId} className="text-sm font-medium">
                Workspace Name
              </label>
              <Input
                id={nameId}
                placeholder="e.g., My Monitoring"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isSubmitting || isLoading}
                required
              />
            </div>
            <div className="grid gap-2">
              <label htmlFor={typeId} className="text-sm font-medium">
                Type
              </label>
              <Select
                value={type}
                onValueChange={(value) => setType(value as WorkspaceType)}
                disabled={isSubmitting || isLoading}
              >
                <SelectTrigger id={typeId}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="Personal">Personal</SelectItem>
                  <SelectItem value="Team">Team</SelectItem>
                  <SelectItem value="Competitor">Competitor</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Tags</label>
              <TagInput tags={tags} onChange={setTags} disabled={isSubmitting || isLoading} />
            </div>
            {error && (
              <div className="rounded-md bg-red-50 p-2 text-sm text-red-700">
                {error.message || 'Something went wrong'}
              </div>
            )}
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting || isLoading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting || isLoading || !name?.trim()}>
              {isSubmitting || isLoading ? 'Saving...' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
