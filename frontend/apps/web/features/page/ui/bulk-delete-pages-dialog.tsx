'use client'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@workspace/ui/components/atoms/alert-dialog'
import { AlertCircle } from 'lucide-react'

export interface BulkDeletePagesDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => Promise<void>
  count: number
  isLoading?: boolean
}

export function BulkDeletePagesDialog({
  open,
  onOpenChange,
  onConfirm,
  count,
  isLoading = false,
}: Readonly<BulkDeletePagesDialogProps>) {
  const handleConfirm = async () => {
    await onConfirm()
    onOpenChange(false)
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <div className="flex items-center gap-2 text-destructive">
            <AlertCircle className="h-5 w-5" />
            <AlertDialogTitle>Delete {count} {count === 1 ? 'Page' : 'Pages'}</AlertDialogTitle>
          </div>
          <AlertDialogDescription>
            Are you sure you want to delete{' '}
            <span className="font-medium text-foreground">
              {count} {count === 1 ? 'page' : 'pages'}
            </span>
            ? This action cannot be undone and will remove all associated monitoring data.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleConfirm}
            disabled={isLoading}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {isLoading ? 'Deleting...' : `Delete ${count === 1 ? 'Page' : `${count} Pages`}`}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
