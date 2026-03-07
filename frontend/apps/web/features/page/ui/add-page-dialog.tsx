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
import {
  Link2,
  Sparkles,
  Loader2,
  ArrowLeft,
  Monitor,
  MousePointerClick,
  AlertTriangle,
  RotateCcw,
} from 'lucide-react'
import { useCallback, useEffect, useId, useRef, useState } from 'react'
import { CHECK_FREQUENCY_OPTIONS } from '../domain/types'
import type { CreatePageDto, PagePreviewResult, SelectorOffsets } from '../domain/types'
import { PagePreviewSelector, type ElementSelection } from './page-preview-selector'
import { PageApi } from '@workspace/services/page-api'

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

type WizardStep = 'url' | 'selector' | 'config'

export interface WorkspaceOption {
  id: string
  name: string
}

export interface AddPageDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: CreatePageDto) => Promise<void>
  workspaceId?: string
  workspaces?: WorkspaceOption[]
  isLoading?: boolean
  error?: Error | null
}

export function AddPageDialog({
  open,
  onOpenChange,
  onSubmit,
  workspaceId,
  workspaces = [],
  isLoading = false,
  error,
}: Readonly<AddPageDialogProps>) {
  const uid = useId()
  const [step, setStep] = useState<WizardStep>('url')
  const [selectedWorkspaceId, setSelectedWorkspaceId] = useState(workspaceId ?? '')
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

  // Selector state
  const [selectorType, setSelectorType] = useState<'full_page' | 'element' | 'sections'>('full_page')
  const [cssSelector, setCssSelector] = useState('')
  const [xpathSelector, setXpathSelector] = useState('')
  const [selectorOffsets, setSelectorOffsets] = useState<SelectorOffsets>({
    top: 0,
    right: 0,
    bottom: 0,
    left: 0,
  })
  const [sections, setSections] = useState<ElementSelection[]>([])
  const [previewLoading, setPreviewLoading] = useState(false)
  const [previewProgress, setPreviewProgress] = useState<string | null>(null)
  const [previewData, setPreviewData] = useState<PagePreviewResult | null>(null)
  const [previewError, setPreviewError] = useState<string | null>(null)
  const previewAbortRef = useRef<AbortController | null>(null)

  useEffect(() => {
    if (!open) {
      // Abort any in-flight preview stream
      previewAbortRef.current?.abort()
      previewAbortRef.current = null
      setStep('url')
      setSelectedWorkspaceId(workspaceId ?? '')
      setName('')
      setUrl('')
      setTags([])
      setCheckFrequency('Off')
      setBlockAdsCookies(true)
      setScheduleType('all_time')
      setEnabledInsightTypes(['marketing', 'market_analysis'])
      setEnabledAlertConditions(['any_changes'])
      setCustomAlertCondition('')
      setSelectorType('full_page')
      setCssSelector('')
      setXpathSelector('')
      setSelectorOffsets({ top: 0, right: 0, bottom: 0, left: 0 })
      setSections([])
      setPreviewLoading(false)
      setPreviewProgress(null)
      setPreviewData(null)
      setPreviewError(null)
    }
  }, [open, workspaceId])

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

  const handlePreviewPage = useCallback(async () => {
    if (!url.trim()) return
    // Abort any existing preview request
    previewAbortRef.current?.abort()
    const controller = new AbortController()
    previewAbortRef.current = controller

    setPreviewLoading(true)
    setPreviewProgress(null)
    setPreviewError(null)
    try {
      const result = await PageApi.previewPageWithProgress(
        url.trim(),
        blockAdsCookies,
        (progress) => setPreviewProgress(progress.message),
        controller.signal,
      )
      setPreviewData(result)
      setStep('selector')
    } catch (err) {
      // Don't show error for user-initiated aborts
      if (err instanceof DOMException && err.name === 'AbortError') return
      setPreviewError(err instanceof Error ? err.message : 'Failed to preview page')
    } finally {
      setPreviewLoading(false)
      setPreviewProgress(null)
    }
  }, [url, blockAdsCookies])

  const handleElementSelect = useCallback((selection: ElementSelection | null) => {
    if (selection) {
      setSelectorType('element')
      setCssSelector(selection.cssSelector)
      setXpathSelector(selection.xpathSelector)
      setSelectorOffsets(selection.offsets)
    } else {
      setSelectorType('full_page')
      setCssSelector('')
      setXpathSelector('')
      setSelectorOffsets({ top: 0, right: 0, bottom: 0, left: 0 })
    }
  }, [])

  const handleMultiSelect = useCallback((selections: ElementSelection[]) => {
    setSections(selections)
    if (selections.length > 0) {
      setSelectorType('sections')
      // Clear legacy single-selector fields
      setCssSelector('')
      setXpathSelector('')
      setSelectorOffsets({ top: 0, right: 0, bottom: 0, left: 0 })
    } else {
      setSelectorType('full_page')
    }
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedWorkspaceId || !name.trim() || !url.trim()) return

    await onSubmit({
      workspaceId: selectedWorkspaceId,
      name: name.trim(),
      url: url.trim(),
      tags,
      checkFrequency,
      blockAdsCookies,
      scheduleType,
      enabledInsightTypes,
      enabledAlertConditions,
      customAlertCondition: customAlertCondition.trim(),
      selectorType,
      cssSelector,
      xpathSelector,
      selectorOffsets,
      sections: selectorType === 'sections'
        ? sections.map((s, i) => ({
          name: s.name || `Section ${i + 1}`,
          cssSelector: s.cssSelector,
          xpathSelector: s.xpathSelector,
          selectorOffsets: s.offsets,
          rect: s.rect,
          viewportWidth: previewData?.viewport.width,
          sortOrder: i,
        }))
        : undefined,
    })
  }

  const isFormValid = selectedWorkspaceId !== '' && name.trim() !== '' && url.trim() !== ''

  const stepTitle = {
    url: 'Add New Page',
    selector: 'Select Region to Monitor',
    config: 'Configure Monitoring',
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className={`${step === 'selector' ? 'w-[95vw] max-w-6xl max-h-[95vh]' : 'max-w-[95vw] sm:max-w-[672px] max-h-[95vh] sm:max-h-[90vh]'} p-0 flex flex-col gap-0 overflow-hidden`}
      >
        {/* Header */}
        <DialogHeader className="px-6 py-5 border-b border-border shrink-0">
          <DialogTitle className="text-xl font-bold flex items-center gap-2">
            {step !== 'url' && (
              <button
                type="button"
                onClick={() => setStep(step === 'config' ? 'selector' : 'url')}
                className="p-1 -ml-1 hover:bg-muted rounded"
              >
                <ArrowLeft className="h-4 w-4" />
              </button>
            )}
            {stepTitle[step]}
          </DialogTitle>
        </DialogHeader>

        <form
          onSubmit={handleSubmit}
          className="flex-1 overflow-hidden flex flex-col min-h-0"
        >
          {/* Scrollable body */}
          <div className={`flex-1 px-6 py-6 min-h-0 ${step === 'selector' ? 'flex flex-col gap-4 overflow-hidden' : 'overflow-y-auto space-y-8'}`}>
            {/* ─── STEP 1: URL Entry ─── */}
            {step === 'url' && (
              <>
                {/* Workspace selector */}
                {!workspaceId && (
                  <div className="space-y-2">
                    <Label className="font-semibold">Workspace</Label>
                    <Select
                      value={selectedWorkspaceId}
                      onValueChange={setSelectedWorkspaceId}
                      disabled={isLoading}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select a workspace" />
                      </SelectTrigger>
                      <SelectContent>
                        {workspaces.map((ws) => (
                          <SelectItem key={ws.id} value={ws.id}>
                            {ws.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                )}

                {/* Page URL */}
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
                      disabled={isLoading || previewLoading}
                      required
                    />
                  </div>
                </div>

                {/* Block ads checkbox */}
                <div className="flex items-center gap-2">
                  <Checkbox
                    id={`${uid}-block-ads-step1`}
                    checked={blockAdsCookies}
                    onCheckedChange={(checked) => setBlockAdsCookies(checked === true)}
                    disabled={isLoading || previewLoading}
                  />
                  <Label htmlFor={`${uid}-block-ads-step1`} className="font-normal cursor-pointer">
                    Block ads and cookie banners
                  </Label>
                </div>

                {previewError && (
                  <div className="flex items-start gap-3 rounded-lg border border-destructive/30 bg-destructive/5 p-4">
                    <AlertTriangle className="h-5 w-5 text-destructive shrink-0 mt-0.5" />
                    <div className="flex-1 space-y-2">
                      <p className="text-sm font-medium text-destructive">Preview failed</p>
                      <p className="text-sm text-muted-foreground">{previewError}</p>
                      <div className="flex gap-2 pt-1">
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={handlePreviewPage}
                          disabled={previewLoading}
                        >
                          <RotateCcw className="h-3.5 w-3.5 mr-1.5" />
                          Try Again
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            setPreviewError(null)
                            setStep('config')
                          }}
                        >
                          Skip Preview
                        </Button>
                      </div>
                    </div>
                  </div>
                )}
              </>
            )}

            {/* ─── STEP 2: Element Selection ─── */}
            {step === 'selector' && previewData && (
              <>
                <p className="text-sm text-muted-foreground shrink-0">
                  Click elements to monitor specific sections, or skip to monitor the full page.
                  You can select multiple sections.
                </p>

                <div className="flex-1 min-h-0 overflow-hidden flex flex-col">
                  <PagePreviewSelector
                    screenshotBase64={previewData.screenshot_base64}
                    viewport={previewData.viewport}
                    pageHeight={previewData.page_height}
                    elements={previewData.elements}
                    onMultiSelect={handleMultiSelect}
                    multiSelect
                  />
                </div>

                {selectorType === 'sections' && sections.length > 0 && (
                  <div className="flex items-center gap-2 text-sm text-blue-600 bg-blue-50 dark:bg-blue-950/30 px-3 py-2 rounded shrink-0">
                    <MousePointerClick className="h-4 w-4 shrink-0" />
                    <span>Monitoring {sections.length} section{sections.length > 1 ? 's' : ''}</span>
                  </div>
                )}

                {selectorType === 'full_page' && (
                  <div className="flex items-center gap-2 text-sm text-muted-foreground bg-muted px-3 py-2 rounded shrink-0">
                    <Monitor className="h-4 w-4 shrink-0" />
                    <span>Monitoring full page (click elements above to select sections)</span>
                  </div>
                )}
              </>
            )}

            {/* ─── STEP 3: Configuration ─── */}
            {step === 'config' && (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 sm:gap-8 items-start">
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
            )}

            {error && <p className="text-sm text-destructive">{error.message}</p>}
          </div>

          {/* Footer — always visible */}
          <DialogFooter className="px-6 py-4 bg-muted border-t border-border shrink-0 flex-row justify-between sm:justify-between sm:space-x-0">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isLoading || previewLoading}
            >
              Cancel
            </Button>

            <div className="flex gap-2">
              {step === 'url' && (
                <>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setStep('config')}
                    disabled={!url.trim() || previewLoading}
                  >
                    Skip Preview
                  </Button>
                  <Button
                    type="button"
                    onClick={handlePreviewPage}
                    disabled={!url.trim() || previewLoading}
                  >
                    {previewLoading ? (
                      <>
                        <Loader2 className="h-4 w-4 animate-spin mr-2" />
                        {previewProgress ?? 'Loading Preview...'}
                      </>
                    ) : (
                      'Preview Page'
                    )}
                  </Button>
                </>
              )}

              {step === 'selector' && (
                <Button type="button" onClick={() => setStep('config')}>
                  Continue
                </Button>
              )}

              {step === 'config' && (
                <Button type="submit" disabled={isLoading || !isFormValid}>
                  {isLoading ? 'Adding...' : '+ Add Page'}
                </Button>
              )}
            </div>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
