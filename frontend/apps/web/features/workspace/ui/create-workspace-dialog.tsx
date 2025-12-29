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
import type { Workspace, WorkspaceType } from '@/features/workspace/domain/types'

export interface CreateWorkspaceDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: { name: string; type: WorkspaceType }) => Promise<void>
  isLoading: boolean
  error: Error | null
  initialData?: Workspace | null
  mode: 'create' | 'edit'
}

export function CreateWorkspaceDialog({
  open,
  onOpenChange,
  onSubmit,
  isLoading,
  error,
  initialData,
  mode,
}: Readonly<CreateWorkspaceDialogProps>) {
  const nameId = useId()
  const typeId = useId()
  const [name, setName] = useState('')
  const [type, setType] = useState<WorkspaceType>('Personal')
  const [isSubmitting, setIsSubmitting] = useState(false)

  // Initialize form with data when mode changes or initialData changes
  useEffect(() => {
    if (initialData && mode === 'edit') {
      setName(initialData.name)
      setType(initialData.type as WorkspaceType)
    } else {
      setName('')
      setType('Personal')
    }
  }, [
    initialData,
    mode,
  ])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return

    setIsSubmitting(true)
    try {
      await onSubmit({
        name: name.trim(),
        type,
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      setName('')
      setType('Personal')
    }
    onOpenChange(newOpen)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>{mode === 'create' ? 'Create Workspace' : 'Edit Workspace'}</DialogTitle>
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
              onClick={() => handleOpenChange(false)}
              disabled={isSubmitting || isLoading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting || isLoading || !name.trim()}>
              {(() => {
                const isLoadingState = isSubmitting || isLoading
                const loadingText = mode === 'create' ? 'Creating...' : 'Updating...'
                const defaultText = mode === 'create' ? 'Create' : 'Update'
                return isLoadingState ? loadingText : defaultText
              })()}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
