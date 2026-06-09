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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@workspace/ui/components/atoms/select'
import { Loader2 } from 'lucide-react'
import { useId, useState } from 'react'
import type { CreateSocialProfileDto, Platform } from '../domain/types'
import { INTERVAL_PRESETS } from '../domain/types'

interface AddSocialProfileDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: CreateSocialProfileDto) => Promise<void>
  workspaceId: string
  isLoading?: boolean
}

export function AddSocialProfileDialog({
  open,
  onOpenChange,
  onSubmit,
  workspaceId,
  isLoading = false,
}: Readonly<AddSocialProfileDialogProps>) {
  const uid = useId()
  const [platform, setPlatform] = useState<Platform>('instagram')
  const [handle, setHandle] = useState('')
  const [intervalMinutes, setIntervalMinutes] = useState<number>(1440)
  const [validationError, setValidationError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setValidationError(null)

    if (!handle.trim()) {
      setValidationError('Handle is required.')
      return
    }

    try {
      await onSubmit({
        workspaceId,
        platform,
        handle: handle.trim().replace(/^@/, ''),
        checkIntervalMinutes: intervalMinutes,
      })
      // reset on success
      setHandle('')
      setPlatform('instagram')
      setIntervalMinutes(1440)
    } catch (err) {
      setValidationError(err instanceof Error ? err.message : 'Failed to add profile.')
    }
  }

  const handleOpenChange = (next: boolean) => {
    if (!next) {
      setHandle('')
      setValidationError(null)
    }
    onOpenChange(next)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Add Social Profile</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4 py-2">
          {/* Platform select — Phase 1: Instagram only */}
          <div className="flex flex-col gap-1.5">
            <Label htmlFor={`${uid}-platform`}>Platform</Label>
            <Select
              value={platform}
              onValueChange={(v) => setPlatform(v as Platform)}
            >
              <SelectTrigger id={`${uid}-platform`}>
                <SelectValue placeholder="Select platform" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="instagram">Instagram</SelectItem>
                <SelectItem value="tiktok" disabled>
                  TikTok (coming soon)
                </SelectItem>
                <SelectItem value="facebook" disabled>
                  Facebook (coming soon)
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Handle input */}
          <div className="flex flex-col gap-1.5">
            <Label htmlFor={`${uid}-handle`}>Username / Handle</Label>
            <Input
              id={`${uid}-handle`}
              placeholder="e.g. nasa (without @)"
              value={handle}
              onChange={(e) => setHandle(e.target.value)}
              autoComplete="off"
            />
          </div>

          {/* Check interval */}
          <div className="flex flex-col gap-1.5">
            <Label htmlFor={`${uid}-interval`}>Check interval</Label>
            <Select
              value={String(intervalMinutes)}
              onValueChange={(v) => setIntervalMinutes(Number(v))}
            >
              <SelectTrigger id={`${uid}-interval`}>
                <SelectValue placeholder="Select interval" />
              </SelectTrigger>
              <SelectContent>
                {INTERVAL_PRESETS.map((preset) => (
                  <SelectItem key={preset.value} value={String(preset.value)}>
                    {preset.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {validationError && (
            <p className="text-sm text-destructive">{validationError}</p>
          )}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
              Add Profile
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
