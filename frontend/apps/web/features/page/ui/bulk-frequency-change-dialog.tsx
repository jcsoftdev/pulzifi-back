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
import { Clock } from 'lucide-react'

export interface BulkFrequencyChangeDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => Promise<void>
  count: number
  frequency: string
  isLoading?: boolean
}

export function BulkFrequencyChangeDialog({
  open,
  onOpenChange,
  onConfirm,
  count,
  frequency,
  isLoading = false,
}: Readonly<BulkFrequencyChangeDialogProps>) {
  const handleConfirm = async () => {
    await onConfirm()
    onOpenChange(false)
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <div className="flex items-center gap-2 text-foreground">
            <Clock className="h-5 w-5" />
            <AlertDialogTitle>Update Check Frequency</AlertDialogTitle>
          </div>
          <AlertDialogDescription>
            Set check frequency to{' '}
            <span className="font-medium text-foreground">{frequency}</span> for{' '}
            <span className="font-medium text-foreground">
              {count} {count === 1 ? 'page' : 'pages'}
            </span>
            ?
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={handleConfirm} disabled={isLoading}>
            {isLoading ? 'Updating...' : 'Confirm'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
