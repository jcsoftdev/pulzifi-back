'use client'

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@workspace/ui/components/atoms/dialog'

interface TeamsConsentModalProps {
  open: boolean
  onClose: () => void
  adminURL: string
}

export function TeamsConsentModal({ open, onClose, adminURL }: Readonly<TeamsConsentModalProps>) {
  const handleCopy = () => {
    navigator.clipboard.writeText(adminURL).catch(() => {
      // Clipboard may not be available in all environments; fail silently
    })
  }

  return (
    <Dialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onClose() }}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Microsoft Teams admin approval required</DialogTitle>
          <DialogDescription>
            Your Microsoft tenant administrator needs to approve this app before users can connect.
            Forward this link to your IT admin:
          </DialogDescription>
        </DialogHeader>

        <div className="rounded-lg bg-muted p-3 text-xs font-mono break-all select-all">
          {adminURL}
        </div>

        <DialogFooter>
          <button
            type="button"
            onClick={onClose}
            className="text-xs font-medium bg-muted hover:bg-muted/70 text-foreground px-4 py-2 rounded-lg transition-colors"
          >
            Close
          </button>
          <button
            type="button"
            onClick={handleCopy}
            className="text-xs font-medium bg-primary hover:bg-primary/90 text-primary-foreground px-4 py-2 rounded-lg transition-colors"
          >
            Copy link
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
