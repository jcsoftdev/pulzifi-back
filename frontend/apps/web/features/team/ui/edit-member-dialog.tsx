'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@workspace/ui/components/atoms/dialog'
import { Label } from '@workspace/ui/components/atoms/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@workspace/ui/components/atoms/select'
import { useEffect, useState } from 'react'
import type { TeamMember } from '../domain/types'
import { memberFullName } from '../domain/types'

interface EditMemberDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  member: TeamMember | null
  onSubmit: (memberId: string, role: string) => Promise<void>
  isLoading: boolean
  error: Error | null
}

export function EditMemberDialog({
  open,
  onOpenChange,
  member,
  onSubmit,
  isLoading,
  error,
}: Readonly<EditMemberDialogProps>) {
  const [role, setRole] = useState<string>(member?.role ?? 'MEMBER')
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (member) setRole(member.role)
  }, [member])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!member) return

    setIsSubmitting(true)
    try {
      await onSubmit(member.id, role)
      onOpenChange(false)
    } finally {
      setIsSubmitting(false)
    }
  }

  const busy = isSubmitting || isLoading

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>Edit member</DialogTitle>
        </DialogHeader>
        {member && (
          <form onSubmit={handleSubmit}>
            <div className="grid gap-4 py-4">
              <p className="text-sm text-muted-foreground">
                Changing role for <span className="font-medium text-foreground">{memberFullName(member)}</span>
              </p>
              <div className="grid gap-2">
                <Label htmlFor="edit-role">Role</Label>
                <Select value={role} onValueChange={setRole} disabled={busy}>
                  <SelectTrigger id="edit-role">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="MEMBER">Member</SelectItem>
                    <SelectItem value="ADMIN">Admin</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              {error && (
                <p className="text-sm text-destructive">{error.message || 'Something went wrong'}</p>
              )}
            </div>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
                disabled={busy}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={busy}>
                {busy ? 'Saving...' : 'Save changes'}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}
