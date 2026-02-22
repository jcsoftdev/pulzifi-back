'use client'

import type { Page } from '@workspace/services'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@workspace/ui/components/atoms/select'
import { useState } from 'react'

interface CreateReportDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: { pageId: string; title: string; reportDate: string }) => Promise<void>
  pages: Page[]
  isLoading: boolean
  error: Error | null
}

export function CreateReportDialog({
  open,
  onOpenChange,
  onSubmit,
  pages,
  isLoading,
  error,
}: Readonly<CreateReportDialogProps>) {
  const [title, setTitle] = useState('')
  const [pageId, setPageId] = useState('')
  const [reportDate, setReportDate] = useState(() => new Date().toISOString().split('T')[0] ?? '')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim() || !pageId) return

    setIsSubmitting(true)
    try {
      await onSubmit({ pageId, title: title.trim(), reportDate })
      setTitle('')
      setPageId('')
      setReportDate(new Date().toISOString().split('T')[0] ?? '')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      setTitle('')
      setPageId('')
      setReportDate(new Date().toISOString().split('T')[0] ?? '')
    }
    onOpenChange(newOpen)
  }

  const busy = isSubmitting || isLoading

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>Create report</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="report-title">Title</Label>
              <Input
                id="report-title"
                type="text"
                placeholder="Monthly report"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                disabled={busy}
                required
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="report-page">Page</Label>
              <Select value={pageId} onValueChange={setPageId} disabled={busy}>
                <SelectTrigger id="report-page">
                  <SelectValue placeholder="Select a page" />
                </SelectTrigger>
                <SelectContent>
                  {pages.map((page) => (
                    <SelectItem key={page.id} value={page.id}>
                      {page.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="report-date">Report date</Label>
              <Input
                id="report-date"
                type="date"
                value={reportDate}
                onChange={(e) => setReportDate(e.target.value)}
                disabled={busy}
                required
              />
            </div>
            {error && (
              <p className="text-sm text-destructive">{error.message || 'Something went wrong'}</p>
            )}
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={busy}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={busy || !title.trim() || !pageId}>
              {busy ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
