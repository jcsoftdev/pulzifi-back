'use client'

import { type MonitoringConfig, type Page, PageApi } from '@workspace/services/page-api'
import { Badge } from '@workspace/ui/components/atoms/badge'
import { Button } from '@workspace/ui/components/atoms/button'
import { Input } from '@workspace/ui/components/atoms/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@workspace/ui/components/atoms/select'
import { Check, Loader2, Pencil, X } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { useState, useTransition } from 'react'
import { CHECK_FREQUENCY_OPTIONS } from '@/features/page/domain/types'
import { notification } from '@/lib/notification'

interface GeneralSummaryCardProps {
  page: Page
  config: MonitoringConfig | null
  quotaExceeded?: boolean
  refillDate?: string
}

export function GeneralSummaryCard({ page, config, quotaExceeded, refillDate }: Readonly<GeneralSummaryCardProps>) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()
  const [isEditingTags, setIsEditingTags] = useState(false)
  const [newTag, setNewTag] = useState('')

  const handleUpdateConfig = (updates: Partial<MonitoringConfig>) => {
    startTransition(async () => {
      try {
        const result = await PageApi.updateMonitoringConfig(page.id, updates)
        router.refresh()
        if (updates.checkFrequency) {
          window.dispatchEvent(new CustomEvent('checks:refresh'))
        }
        if (result.quotaExceeded) {
          notification.warning({
            title: 'Monthly check quota reached',
            description: `Settings saved, but checks are paused until ${refillDate ?? 'next billing cycle'}.`,
          })
        } else {
          notification.success({
            title: 'Settings saved',
          })
        }
      } catch (error) {
        console.error('Failed to update config', error)
        notification.error({
          title: 'Failed to save settings',
          description: error instanceof Error ? error.message : 'Please try again.',
        })
      }
    })
  }

  const handleUpdatePage = (
    updates: Partial<Page> & {
      tags?: string[]
    }
  ) => {
    startTransition(async () => {
      try {
        await PageApi.updatePage(page.id, updates)
        router.refresh()
        setIsEditingTags(false)
        notification.success({
          title: 'Page updated',
        })
      } catch (error) {
        console.error('Failed to update page', error)
        notification.error({
          title: 'Failed to update page',
          description: error instanceof Error ? error.message : 'Please try again.',
        })
      }
    })
  }

  const handleAddTag = () => {
    if (!newTag.trim()) return
    const currentTags = page.tags || []
    if (currentTags.includes(newTag.trim())) return

    handleUpdatePage({
      tags: [
        ...currentTags,
        newTag.trim(),
      ],
    })
    setNewTag('')
  }

  const handleRemoveTag = (tagToRemove: string) => {
    const currentTags = page.tags || []
    handleUpdatePage({
      tags: currentTags.filter((tag) => tag !== tagToRemove),
    })
  }

  return (
    <div className="flex flex-col gap-6 bg-card border border-border rounded-xl p-6 h-full">
      <div className="flex items-center justify-between">
        <h3 className="text-xl font-semibold text-foreground">General Summary</h3>
        {isPending && <Loader2 className="w-4 h-4 animate-spin text-muted-foreground" />}
      </div>

      {quotaExceeded && (
        <div className="rounded-md border border-yellow-500/30 bg-yellow-500/10 px-4 py-3 text-sm text-yellow-700 dark:text-yellow-400">
          Monthly check quota reached. Checks are paused until {refillDate ?? 'next billing cycle'}.
        </div>
      )}

      <div className="flex flex-col gap-4">
        {/* Tag Section */}
        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-muted-foreground">Tag</span>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              onClick={() => setIsEditingTags(!isEditingTags)}
              aria-label="Edit tags"
              className="text-muted-foreground hover:text-foreground"
            >
              <Pencil className="w-4 h-4" />
            </Button>
          </div>

          <div className="flex flex-wrap gap-2">
            {page.tags?.map((tag) => (
              <div key={tag} className="flex items-center gap-1">
                <Badge variant="outline" className="px-2 py-0.5 text-xs">
                  {tag}
                </Badge>
                {isEditingTags && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-sm"
                    onClick={() => handleRemoveTag(tag)}
                    aria-label={`Remove ${tag}`}
                    className="hover:text-destructive"
                  >
                    <X className="w-3 h-3" />
                  </Button>
                )}
              </div>
            ))}

            {isEditingTags && (
              <div className="flex items-center gap-2">
                <Input
                  type="text"
                  value={newTag}
                  onChange={(e) => setNewTag(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleAddTag()}
                  placeholder="Add tag..."
                  className="h-7 px-2 text-xs w-28"
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  onClick={handleAddTag}
                  aria-label="Add tag"
                  className="hover:text-primary"
                >
                  <Check className="w-4 h-4" />
                </Button>
              </div>
            )}

            {(!page.tags || page.tags.length === 0) && !isEditingTags && (
              <span className="text-sm text-muted-foreground italic">No tags</span>
            )}
          </div>
        </div>

        {/* Check Frequency */}
        <div className="flex flex-col gap-2">
          <span className="text-sm font-medium text-muted-foreground">Check Frequency</span>
          <Select
            value={config?.checkFrequency || page.checkFrequency || 'Off'}
            onValueChange={(value) =>
              handleUpdateConfig({
                checkFrequency: value,
              })
            }
          >
            <SelectTrigger className="h-7 text-sm border-none shadow-none hover:bg-muted/50 focus:ring-0">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {CHECK_FREQUENCY_OPTIONS.map((freq) => (
                <SelectItem key={freq.value} value={freq.value}>
                  {freq.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Options List */}
        <div className="flex flex-col gap-3 mt-2">
          <button
            type="button"
            className="flex items-center gap-2 cursor-pointer group"
            onClick={() =>
              handleUpdateConfig({
                blockAdsCookies: !config?.blockAdsCookies,
              })
            }
          >
            <div
              className={`flex items-center justify-center w-4 h-4 rounded-sm transition-colors ${config?.blockAdsCookies ? 'bg-destructive' : 'border border-muted-foreground'}`}
            >
              {config?.blockAdsCookies && <Check className="w-3 h-3 text-white" />}
            </div>
            <span className="text-sm text-foreground group-hover:text-foreground/80 transition-colors">
              Block ads and cookie banners
            </span>
          </button>
        </div>
      </div>
    </div>
  )
}
