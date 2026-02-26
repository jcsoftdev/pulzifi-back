'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { Checkbox } from '@workspace/ui/components/atoms/checkbox'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@workspace/ui/components/atoms/dialog'
import { Input } from '@workspace/ui/components/atoms/input'
import { Label } from '@workspace/ui/components/atoms/label'
import { TagsInput } from '@workspace/ui/components/atoms/tags-input'
import { RadioGroup, RadioGroupItem } from '@workspace/ui/components/atoms/radio-group'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@workspace/ui/components/atoms/select'
import { Textarea } from '@workspace/ui/components/atoms/textarea'
import { Link2, Sparkles } from 'lucide-react'
import { useEffect, useId, useState } from 'react'
import { CHECK_FREQUENCY_OPTIONS } from '../domain/types'
import type { CreatePageDto } from '../domain/types'

const INSIGHT_TYPES = [
  { value: 'marketing', label: 'Marketing Lens' },
  { value: 'market_analysis', label: 'Market Analysis' },
  { value: 'business_opportunities', label: 'Business Opportunities' },
  { value: 'job_recommendation', label: 'Job recommendation' },
] as const

const ALERT_CONDITIONS = [
  { value: 'any_changes', label: 'Any changes' },
  { value: 'navigation_changes', label: "Site's main navigation menu changes" },
] as const

export interface AddPageDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: CreatePageDto) => Promise<void>
  workspaceId: string
  isLoading?: boolean
  error?: Error | null
}

export function AddPageDialog({
  open,
  onOpenChange,
  onSubmit,
  workspaceId,
  isLoading = false,
  error,
}: Readonly<AddPageDialogProps>) {
  const uid = useId()
  const [name, setName] = useState('')
  const [url, setUrl] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [checkFrequency, setCheckFrequency] = useState('Off')
  const [blockAdsCookies, setBlockAdsCookies] = useState(true)
  const [scheduleType, setScheduleType] = useState('all_time')
  const [enabledInsightTypes, setEnabledInsightTypes] = useState<string[]>([
    'marketing',
    'market_analysis',
  ])
  const [enabledAlertConditions, setEnabledAlertConditions] = useState<string[]>(['any_changes'])
  const [customAlertCondition, setCustomAlertCondition] = useState('')

  useEffect(() => {
    if (!open) {
      setName('')
      setUrl('')
      setTags([])
      setCheckFrequency('Off')
      setBlockAdsCookies(true)
      setScheduleType('all_time')
      setEnabledInsightTypes(['marketing', 'market_analysis'])
      setEnabledAlertConditions(['any_changes'])
      setCustomAlertCondition('')
    }
  }, [open])

  const toggleInsightType = (value: string, checked: boolean) => {
    setEnabledInsightTypes((prev) =>
      checked ? [...prev, value] : prev.filter((t) => t !== value)
    )
  }

  const toggleAlertCondition = (value: string, checked: boolean) => {
    setEnabledAlertConditions((prev) =>
      checked ? [...prev, value] : prev.filter((c) => c !== value)
    )
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim() || !url.trim()) return

    await onSubmit({
      workspaceId,
      name: name.trim(),
      url: url.trim(),
      tags,
      checkFrequency,
      blockAdsCookies,
      scheduleType,
      enabledInsightTypes,
      enabledAlertConditions,
      customAlertCondition: customAlertCondition.trim(),
    })
  }

  const isFormValid = name.trim() !== '' && url.trim() !== ''

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[672px] max-h-[90vh] p-0 flex flex-col gap-0 overflow-hidden">
        {/* Header */}
        <DialogHeader className="px-6 py-5 border-b border-border shrink-0">
          <DialogTitle className="text-xl font-bold">Add New Page</DialogTitle>
        </DialogHeader>

        <form
          onSubmit={handleSubmit}
          className="flex-1 overflow-hidden flex flex-col min-h-0"
        >
          {/* Scrollable body */}
          <div className="flex-1 overflow-y-auto px-6 py-6 space-y-8">
            {/* Page URL — full width */}
            <div className="space-y-2">
              <Label className="font-semibold">Page URL</Label>
              <div className="relative">
                <Link2 className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
                <Input
                  type="url"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  placeholder="https://example.com/page-to-monitor"
                  className="pl-9"
                  disabled={isLoading}
                  required
                />
              </div>
            </div>

            {/* Two-column layout */}
            <div className="grid grid-cols-2 gap-8 items-start">
              {/* Left column: Settings */}
              <div className="space-y-6">
                {/* Page Name */}
                <div className="space-y-1.5">
                  <Label>Page Name</Label>
                  <Input
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="e.g. Pricing Page"
                    disabled={isLoading}
                    required
                  />
                </div>

                {/* Tags */}
                <div className="space-y-1.5">
                  <Label className="text-muted-foreground text-xs">Tags</Label>
                  <TagsInput
                    value={tags}
                    onChange={setTags}
                    placeholder="Add tag…"
                    disabled={isLoading}
                  />
                </div>

                {/* Check Frequency */}
                <div className="space-y-2">
                  <Label>Check Frequency</Label>
                  <Select
                    value={checkFrequency}
                    onValueChange={setCheckFrequency}
                    disabled={isLoading}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {CHECK_FREQUENCY_OPTIONS.map((opt) => (
                        <SelectItem key={opt.value} value={opt.value}>
                          {opt.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>

                  {/* Block ads & cookie banners */}
                  <div className="flex items-center gap-2 pt-1">
                    <Checkbox
                      id={`${uid}-block-ads`}
                      checked={blockAdsCookies}
                      onCheckedChange={(checked) => setBlockAdsCookies(checked === true)}
                      disabled={isLoading}
                    />
                    <Label htmlFor={`${uid}-block-ads`} className="font-normal cursor-pointer">
                      Block ads and cookie banners
                    </Label>
                  </div>
                </div>

                {/* Monitoring Schedule */}
                <div className="space-y-2">
                  <Label className="font-semibold">Monitoring Schedule</Label>
                  <RadioGroup
                    value={scheduleType}
                    onValueChange={setScheduleType}
                    disabled={isLoading}
                    className="gap-2"
                  >
                    <div className="flex items-center gap-2">
                      <RadioGroupItem value="work_days" id={`${uid}-sched-work-days`} />
                      <Label htmlFor={`${uid}-sched-work-days`} className="font-normal cursor-pointer">
                        Work days only
                      </Label>
                    </div>
                    <div className="flex items-center gap-2">
                      <RadioGroupItem value="work_days_work_hours" id={`${uid}-sched-work-hours`} />
                      <Label htmlFor={`${uid}-sched-work-hours`} className="font-normal cursor-pointer">
                        Work days, during work hours
                      </Label>
                    </div>
                    <div className="flex items-center gap-2">
                      <RadioGroupItem value="all_time" id={`${uid}-sched-all-time`} />
                      <Label htmlFor={`${uid}-sched-all-time`} className="font-normal cursor-pointer">
                        Every day (24/7)
                      </Label>
                    </div>
                  </RadioGroup>
                </div>
              </div>

              {/* Right column: Intelligent Insights */}
              <div className="border border-primary/20 rounded-xl shadow-[0_0_15px_0_rgba(0,0,0,0.05)] dark:shadow-none overflow-hidden">
                {/* Panel header */}
                <div className="bg-muted border-b border-border px-4 py-4 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Sparkles className="h-4 w-4 text-primary" />
                    <span className="text-base font-bold">Intelligent Insights</span>
                  </div>
                  <span className="border border-primary text-primary text-[10px] font-medium px-2 py-0.5 rounded">
                    AI Powered
                  </span>
                </div>

                {/* Panel body */}
                <div className="p-4 space-y-8">
                  {/* Insight types */}
                  <div className="space-y-3">
                    {INSIGHT_TYPES.map((insight) => (
                      <div key={insight.value} className="flex items-center gap-2">
                        <Checkbox
                          id={`${uid}-insight-${insight.value}`}
                          checked={enabledInsightTypes.includes(insight.value)}
                          onCheckedChange={(checked) =>
                            toggleInsightType(insight.value, checked === true)
                          }
                          disabled={isLoading}
                        />
                        <Label
                          htmlFor={`${uid}-insight-${insight.value}`}
                          className="font-normal cursor-pointer"
                        >
                          {insight.label}
                        </Label>
                      </div>
                    ))}
                    <p className="text-xs font-semibold text-muted-foreground">
                      New insights coming soon...
                    </p>
                  </div>

                  {/* Alert conditions */}
                  <div className="space-y-3">
                    <p className="text-xs font-medium text-muted-foreground">Alert me when</p>
                    {ALERT_CONDITIONS.map((condition) => (
                      <div key={condition.value} className="flex items-center gap-2">
                        <Checkbox
                          id={`${uid}-alert-${condition.value}`}
                          checked={enabledAlertConditions.includes(condition.value)}
                          onCheckedChange={(checked) =>
                            toggleAlertCondition(condition.value, checked === true)
                          }
                          disabled={isLoading}
                        />
                        <Label
                          htmlFor={`${uid}-alert-${condition.value}`}
                          className="font-normal cursor-pointer"
                        >
                          {condition.label}
                        </Label>
                      </div>
                    ))}

                    {/* Custom alert */}
                    <Textarea
                      value={customAlertCondition}
                      onChange={(e) => setCustomAlertCondition(e.target.value)}
                      placeholder="Add your own here"
                      className="min-h-[60px] text-xs bg-accent border-accent-foreground/30 placeholder:text-muted-foreground resize-none"
                      disabled={isLoading}
                    />
                  </div>
                </div>
              </div>
            </div>

            {error && <p className="text-sm text-destructive">{error.message}</p>}
          </div>

          {/* Footer — always visible */}
          <DialogFooter className="px-6 py-4 bg-muted border-t border-border shrink-0 flex-row justify-between sm:justify-between sm:space-x-0">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading || !isFormValid}>
              {isLoading ? 'Adding...' : '+ Add Page'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
