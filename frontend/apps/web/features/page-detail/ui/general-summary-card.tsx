'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { PageApi, type Page, type MonitoringConfig } from '@workspace/services/page-api'
import { ChevronDown, Check, Pencil, Loader2, X } from 'lucide-react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@workspace/ui/components/atoms/select'

interface GeneralSummaryCardProps {
  page: Page
  config: MonitoringConfig | null
}

export function GeneralSummaryCard({ page, config }: Readonly<GeneralSummaryCardProps>) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()
  const [isEditingTags, setIsEditingTags] = useState(false)
  const [newTag, setNewTag] = useState('')

  const handleUpdateConfig = (updates: Partial<MonitoringConfig>) => {
    startTransition(async () => {
      try {
        await PageApi.updateMonitoringConfig(page.id, updates)
        router.refresh()
      } catch (error) {
        console.error('Failed to update config', error)
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
      } catch (error) {
        console.error('Failed to update page', error)
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

  const FREQUENCIES = [
    'Off',
    'Every 1 hour',
    'Every 2 hours',
    'Every 8 hours',
    'Every day',
    'Every 48 hours',
  ] as const

  return (
    <div className="flex flex-col gap-6 bg-card border border-border rounded-xl p-6 h-full">
      <div className="flex items-center justify-between">
        <h3 className="text-xl font-semibold text-foreground">General Summary</h3>
        {isPending && <Loader2 className="w-4 h-4 animate-spin text-muted-foreground" />}
      </div>

      <div className="flex flex-col gap-4">
        {/* Tag Section */}
        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-muted-foreground">Tag</span>
            <button
              onClick={() => setIsEditingTags(!isEditingTags)}
              className="p-1 text-muted-foreground hover:text-foreground transition-colors"
            >
              <Pencil className="w-4 h-4" />
            </button>
          </div>

          <div className="flex flex-wrap gap-2">
            {page.tags &&
              page.tags.map((tag) => (
                <div
                  key={tag}
                  className="flex items-center gap-2 px-3 py-1.5 rounded-md border border-destructive bg-destructive/10"
                >
                  <span className="text-sm text-foreground">{tag}</span>
                  {isEditingTags && (
                    <button
                      onClick={() => handleRemoveTag(tag)}
                      className="hover:text-destructive transition-colors"
                    >
                      <X className="w-3 h-3" />
                    </button>
                  )}
                </div>
              ))}

            {isEditingTags && (
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  value={newTag}
                  onChange={(e) => setNewTag(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleAddTag()}
                  placeholder="Add tag..."
                  className="px-2 py-1 text-sm bg-transparent border border-border rounded-md focus:outline-none focus:border-primary w-24"
                />
                <button onClick={handleAddTag} className="p-1 hover:text-primary transition-colors">
                  <Check className="w-4 h-4" />
                </button>
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
            value={config?.checkFrequency || page.checkFrequency || 'Every day'}
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
              {FREQUENCIES.map((freq) => (
                <SelectItem key={freq} value={freq}>
                  {freq}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Options List */}
        <div className="flex flex-col gap-3 mt-2">
          <button
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
