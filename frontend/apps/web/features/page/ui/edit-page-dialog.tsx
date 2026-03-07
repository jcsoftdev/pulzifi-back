'use client'

import { PageApi } from '@workspace/services/page-api'
import type { MonitoredSection } from '@workspace/services/page-api'
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
import { RadioGroup, RadioGroupItem } from '@workspace/ui/components/atoms/radio-group'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@workspace/ui/components/atoms/select'
import { TagsInput } from '@workspace/ui/components/atoms/tags-input'
import { Textarea } from '@workspace/ui/components/atoms/textarea'
import {
  ArrowLeft,
  Link2,
  Loader2,
  Monitor,
  MousePointerClick,
  Sparkles,
  Layers,
} from 'lucide-react'
import { useCallback, useEffect, useId, useRef, useState } from 'react'
import { CHECK_FREQUENCY_OPTIONS } from '../domain/types'
import type { EditPageDto, Page, PagePreviewResult, SelectorOffsets } from '../domain/types'
import { PagePreviewSelector, type ElementSelection } from './page-preview-selector'

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

type EditStep = 'settings' | 'selector'

export interface EditPageDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (pageId: string, data: EditPageDto) => Promise<void>
  page: Page | null
  isLoading?: boolean
  error?: Error | null
}

export function EditPageDialog({
  open,
  onOpenChange,
  onSubmit,
  page,
  isLoading = false,
  error,
}: Readonly<EditPageDialogProps>) {
  const uid = useId()
  const [step, setStep] = useState<EditStep>('settings')
  const [name, setName] = useState('')
  const [url, setUrl] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [checkFrequency, setCheckFrequency] = useState('24h')
  const [blockAdsCookies, setBlockAdsCookies] = useState(true)
  const [scheduleType, setScheduleType] = useState('all_time')
  const [enabledInsightTypes, setEnabledInsightTypes] = useState<string[]>([
    'marketing',
    'market_analysis',
  ])
  const [enabledAlertConditions, setEnabledAlertConditions] = useState<string[]>(['any_changes'])
  const [customAlertCondition, setCustomAlertCondition] = useState('')
  const [isLoadingConfig, setIsLoadingConfig] = useState(false)

  // Section / selector state
  const [selectorType, setSelectorType] = useState<'full_page' | 'element' | 'sections'>('full_page')
  const [cssSelector, setCssSelector] = useState('')
  const [xpathSelector, setXpathSelector] = useState('')
  const [selectorOffsets, setSelectorOffsets] = useState<SelectorOffsets>({
    top: 0, right: 0, bottom: 0, left: 0,
  })
  const [existingSections, setExistingSections] = useState<MonitoredSection[]>([])
  const [sections, setSections] = useState<ElementSelection[]>([])
  const [sectionsModified, setSectionsModified] = useState(false)

  // Preview state
  const [previewLoading, setPreviewLoading] = useState(false)
  const [previewProgress, setPreviewProgress] = useState<string | null>(null)
  const [previewData, setPreviewData] = useState<PagePreviewResult | null>(null)
  const [previewError, setPreviewError] = useState<string | null>(null)
  const previewAbortRef = useRef<AbortController | null>(null)

  useEffect(() => {
    if (open && page) {
      // Abort any in-flight preview stream
      previewAbortRef.current?.abort()
      previewAbortRef.current = null
      setStep('settings')
      setName(page.name)
      setUrl(page.url)
      setTags(page.tags ?? [])
      setCheckFrequency(page.checkFrequency || 'Off')
      setBlockAdsCookies(true)
      setScheduleType('all_time')
      setEnabledInsightTypes(['marketing', 'market_analysis'])
      setEnabledAlertConditions(['any_changes'])
      setCustomAlertCondition('')
      setSelectorType('full_page')
      setCssSelector('')
      setXpathSelector('')
      setSelectorOffsets({ top: 0, right: 0, bottom: 0, left: 0 })
      setExistingSections([])
      setSections([])
      setSectionsModified(false)
      setPreviewLoading(false)
      setPreviewProgress(null)
      setPreviewData(null)
      setPreviewError(null)

      setIsLoadingConfig(true)
      Promise.all([
        PageApi.getMonitoringConfig(page.id),
        PageApi.listSections(page.id),
      ])
        .then(([config, sectionsList]) => {
          if (config) {
            setCheckFrequency(config.checkFrequency)
            setBlockAdsCookies(config.blockAdsCookies)
            setScheduleType(config.scheduleType)
            setEnabledInsightTypes(config.enabledInsightTypes)
            setEnabledAlertConditions(config.enabledAlertConditions)
            setCustomAlertCondition(config.customAlertCondition)
            setSelectorType(
              (config.selectorType as 'full_page' | 'element' | 'sections') ?? 'full_page'
            )
            setCssSelector(config.cssSelector ?? '')
            setXpathSelector(config.xpathSelector ?? '')
            setSelectorOffsets(
              config.selectorOffsets ?? { top: 0, right: 0, bottom: 0, left: 0 }
            )
          }
          if (sectionsList && sectionsList.length > 0) {
            setExistingSections(sectionsList)
            setSelectorType('sections')
          }
        })
        .finally(() => setIsLoadingConfig(false))
    }
  }, [open, page])

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

  const handleMultiSelect = useCallback((selections: ElementSelection[]) => {
    setSections(selections)
    setSectionsModified(true)
    if (selections.length > 0) {
      setSelectorType('sections')
      setCssSelector('')
      setXpathSelector('')
      setSelectorOffsets({ top: 0, right: 0, bottom: 0, left: 0 })
    } else {
      setSelectorType('full_page')
    }
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!page || !name.trim() || !url.trim()) return

    const data: EditPageDto = {
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
    }

    // Include sections if they were modified via the selector UI
    if (sectionsModified) {
      data.sections = selectorType === 'sections'
        ? sections.map((s, i) => ({
            name: s.name || `Section ${i + 1}`,
            cssSelector: s.cssSelector,
            xpathSelector: s.xpathSelector,
            selectorOffsets: s.offsets,
            rect: s.rect,
            viewportWidth: previewData?.viewport.width,
            sortOrder: i,
          }))
        : [] // Clear sections if user switched away from sections mode
    }

    await onSubmit(page.id, data)
  }

  const isFormValid = name.trim() !== '' && url.trim() !== ''
  const busy = isLoading || isLoadingConfig

  // Summary of current monitoring mode
  const monitoringSummary = () => {
    if (selectorType === 'sections') {
      const count = sectionsModified ? sections.length : existingSections.length
      return `${count} section${count !== 1 ? 's' : ''}`
    }
    if (selectorType === 'element') {
      return 'Single element'
    }
    return 'Full page'
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className={`${step === 'selector' ? 'w-[95vw] max-w-6xl max-h-[95vh]' : 'max-w-[95vw] sm:max-w-[672px] max-h-[95vh] sm:max-h-[90vh]'} p-0 flex flex-col gap-0 overflow-hidden`}
      >
        {/* Header */}
        <DialogHeader className="px-6 py-5 border-b border-border shrink-0">
          <DialogTitle className="text-xl font-bold flex items-center gap-2">
            {step === 'selector' && (
              <button
                type="button"
                onClick={() => setStep('settings')}
                className="p-1 -ml-1 hover:bg-muted rounded"
              >
                <ArrowLeft className="h-4 w-4" />
              </button>
            )}
            {step === 'settings' ? 'Edit Page' : 'Edit Monitored Sections'}
          </DialogTitle>
        </DialogHeader>

        <form
          onSubmit={handleSubmit}
          className="flex-1 overflow-hidden flex flex-col min-h-0"
        >
          {/* Scrollable body */}
          <div className={`flex-1 px-6 py-6 min-h-0 ${step === 'selector' ? 'flex flex-col gap-4 overflow-hidden' : 'overflow-y-auto space-y-8'}`}>
            {/* ─── SETTINGS STEP ─── */}
            {step === 'settings' && (
              <>
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
                      disabled={busy}
                      required
                    />
                  </div>
                </div>

                {/* Monitored Sections Card */}
                <div className="border border-border rounded-lg overflow-hidden">
                  <div className="bg-muted/50 px-4 py-3 flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Layers className="h-4 w-4 text-muted-foreground" />
                      <span className="text-sm font-medium">Monitored Region</span>
                      <span className="text-xs text-muted-foreground bg-muted px-2 py-0.5 rounded">
                        {monitoringSummary()}
                      </span>
                    </div>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={handlePreviewPage}
                      disabled={busy || previewLoading || !url.trim()}
                    >
                      {previewLoading ? (
                        <>
                          <Loader2 className="h-3 w-3 animate-spin mr-1.5" />
                          {previewProgress ?? 'Loading...'}
                        </>
                      ) : (
                        'Edit Sections'
                      )}
                    </Button>
                  </div>

                  {/* Show existing sections summary */}
                  {selectorType === 'sections' && existingSections.length > 0 && !sectionsModified && (
                    <div className="px-4 py-3 border-t border-border space-y-1.5">
                      {existingSections.map((section, i) => (
                        <div
                          key={section.id}
                          className="flex items-center gap-2 text-xs text-muted-foreground"
                        >
                          <span
                            className="w-4 h-4 rounded-full text-white text-[9px] font-bold flex items-center justify-center shrink-0"
                            style={{
                              backgroundColor: [
                                'rgb(59 130 246)', 'rgb(16 185 129)', 'rgb(245 158 11)',
                                'rgb(239 68 68)', 'rgb(168 85 247)', 'rgb(236 72 153)',
                                'rgb(6 182 212)', 'rgb(249 115 22)',
                              ][i % 8],
                            }}
                          >
                            {i + 1}
                          </span>
                          <span className="font-medium text-foreground">{section.name}</span>
                          <span className="font-mono truncate max-w-[200px]" title={section.cssSelector}>
                            {section.cssSelector}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}

                  {/* Show modified sections summary */}
                  {sectionsModified && sections.length > 0 && (
                    <div className="px-4 py-3 border-t border-border">
                      <div className="flex items-center gap-2 text-sm text-blue-600 dark:text-blue-400">
                        <MousePointerClick className="h-4 w-4 shrink-0" />
                        <span>
                          {sections.length} section{sections.length !== 1 ? 's' : ''} selected
                          <span className="text-xs text-muted-foreground ml-1">(modified)</span>
                        </span>
                      </div>
                    </div>
                  )}

                  {/* Full page mode indicator */}
                  {selectorType === 'full_page' && !sectionsModified && existingSections.length === 0 && (
                    <div className="px-4 py-3 border-t border-border">
                      <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <Monitor className="h-4 w-4 shrink-0" />
                        <span>Monitoring full page</span>
                      </div>
                    </div>
                  )}

                  {previewError && (
                    <div className="px-4 py-2 border-t border-border">
                      <p className="text-sm text-destructive">{previewError}</p>
                    </div>
                  )}
                </div>

                {/* Two-column layout (stacks on small screens) */}
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
                        disabled={busy}
                        required
                      />
                    </div>

                    {/* Tags */}
                    <div className="space-y-1.5">
                      <Label className="text-muted-foreground text-xs">Tags</Label>
                      <TagsInput
                        value={tags}
                        onChange={setTags}
                        placeholder="Add tag..."
                        disabled={busy}
                      />
                    </div>

                    {/* Check Frequency */}
                    <div className="space-y-2">
                      <Label>Check Frequency</Label>
                      <Select
                        value={checkFrequency}
                        onValueChange={setCheckFrequency}
                        disabled={busy}
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
                          disabled={busy}
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
                        disabled={busy}
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
                              disabled={busy}
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
                              disabled={busy}
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
                          disabled={busy}
                        />
                      </div>
                    </div>
                  </div>
                </div>
              </>
            )}

            {/* ─── SELECTOR STEP ─── */}
            {step === 'selector' && previewData && (
              <>
                <p className="text-sm text-muted-foreground shrink-0">
                  Click elements to monitor specific sections, or deselect all to monitor the full page.
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

            {error && <p className="text-sm text-destructive">{error.message}</p>}
          </div>

          {/* Footer — always visible */}
          <DialogFooter className="px-6 py-4 bg-muted border-t border-border shrink-0 flex-row justify-between sm:justify-between sm:space-x-0">
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                if (step === 'selector') {
                  setStep('settings')
                } else {
                  onOpenChange(false)
                }
              }}
              disabled={busy}
            >
              {step === 'selector' ? 'Back' : 'Cancel'}
            </Button>

            <div className="flex gap-2">
              {step === 'selector' && (
                <Button type="button" onClick={() => setStep('settings')}>
                  Done
                </Button>
              )}
              {step === 'settings' && (
                <Button type="submit" disabled={busy || !isFormValid}>
                  {isLoading ? 'Saving...' : 'Save Changes'}
                </Button>
              )}
            </div>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
