'use client'

import type { Check, Insight, MonitoredSection, Page, BlockDiffDto } from '@workspace/services/page-api'
import { PageApi, UsageApi } from '@workspace/services'
import { Button } from '@workspace/ui/components/atoms/button'
import { Settings, Zap } from 'lucide-react'
import { useParams, useSearchParams } from 'next/navigation'
import { useEffect, useMemo, useState } from 'react'
import { ChangesViewService } from '@/features/changes-view/domain/changes-view-service'
import { ChangesViewLayout } from '@/features/changes-view/ui/changes-view-layout'
import { IntelligentInsights } from '@/features/changes-view/ui/intelligent-insights'
import { SectionNavTabs } from '@/features/changes-view/ui/section-nav-tabs'
import { TextChanges, type TextChangeSection } from '@/features/changes-view/ui/text-changes'
import { VisualPulse } from '@/features/changes-view/ui/visual-pulse'
import type { DiffRow } from '@/features/changes-view/utils/simple-diff'
import { diffWords } from '@/features/changes-view/utils/simple-diff'

function VisualPulseSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <div className="relative w-full select-none overflow-hidden rounded-lg border border-border shadow-sm bg-muted/10 h-[500px] animate-pulse">
        <div
          className="absolute top-0 bottom-0 w-1 bg-primary z-10 flex items-center justify-center"
          style={{ left: '50%', transform: 'translateX(-50%)' }}
        >
          <div className="w-8 h-16 bg-primary rounded-lg flex items-center justify-center shadow-lg">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-primary-foreground" aria-hidden="true">
              <path d="m9 18 6-6-6-6" />
            </svg>
          </div>
        </div>
      </div>
    </div>
  )
}

/** Convert a pre-computed BlockDiff (from the backend) into a DiffRow for the TextChanges UI. */
function blockDiffToDiffRow(d: BlockDiffDto): DiffRow {
  switch (d.op) {
    case 'added':
      return { kind: 'added', segments: [{ type: 'added', text: d.block.text }] }
    case 'removed':
      return { kind: 'removed', segments: [{ type: 'removed', text: d.block.text }] }
    case 'changed': {
      const oldText = d.old_block?.text ?? ''
      const segments = diffWords(oldText, d.block.text)
      return { kind: 'inline', segments }
    }
  }
}

function InsightsSkeleton() {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <div className="lg:col-span-2 flex flex-col gap-6">
        <div className="bg-card border border-border rounded-xl p-16 flex flex-col items-center justify-center text-center gap-4">
          <Zap className="h-14 w-14 text-violet-500" strokeWidth={1.5} />
          <p className="text-sm text-muted-foreground">New Intelligent Insight available</p>
          <h3 className="text-lg font-bold text-foreground">
            Do you need Intelligent Insights for this change?
          </h3>
          <p className="text-sm text-muted-foreground max-w-sm">
            Pulzify detected a meaningful update and turned it into an actionable insight.
            Click below to see what&apos;s behind the change.
          </p>
          <Button className="mt-2 gap-2 bg-violet-600 hover:bg-violet-700 text-white" disabled>
            <Zap className="h-4 w-4" />
            View Intelligent Insights
          </Button>
        </div>
      </div>

      <div className="lg:col-span-1">
        <div className="flex flex-col gap-6">
          <Button className="w-full gap-2" variant="outline" disabled>
            <Settings className="h-4 w-4" />
            Edit Intelligent Insights
          </Button>
          <div className="bg-muted/30 rounded-lg p-6 flex flex-col items-center justify-center text-center gap-4 border border-border border-dashed">
            <div className="relative w-full aspect-[16/9] rounded-xl overflow-hidden bg-muted shadow-md border border-border animate-pulse" />
            <div className="h-4 w-48 bg-muted rounded animate-pulse" />
            <Button variant="outline" className="bg-background" disabled>
              View Site
            </Button>
            <div className="h-3 w-56 bg-muted rounded animate-pulse" />
          </div>
        </div>
      </div>
    </div>
  )
}

export default function ChangesPage() {
  const params = useParams()
  const searchParams = useSearchParams()
  const workspaceId = params.id as string
  const pageId = params.pageId as string
  const checkIdParam = searchParams.get('checkId')

  const [checks, setChecks] = useState<Check[]>([])
  const [insights, setInsights] = useState<Insight[]>([])
  const [page, setPage] = useState<Page | null>(null)
  const [sections, setSections] = useState<MonitoredSection[]>([])
  const [selectedSectionId, setSelectedSectionId] = useState<string>('all')
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('visual')
  const [storagePeriodDays, setStoragePeriodDays] = useState(7)

  const hasSections = sections.length > 0

  // All section checks (with sectionId) and all parent/full-page checks (without).
  const sectionChecks = useMemo(() => checks.filter((c) => !!c.sectionId), [checks])
  const parentChecks = useMemo(() => checks.filter((c) => !c.sectionId), [checks])

  // Build a list of unique "execution groups" from parent checks for the "All sections" dropdown.
  // Each parent check represents one monitoring execution.
  // For specific section mode, use that section's checks directly.
  const filteredChecks = selectedSectionId === 'all'
    ? (hasSections ? parentChecks : checks)
    : sectionChecks.filter((c) => c.sectionId === selectedSectionId)

  // Checks within the storage period that appear in the dropdown.
  const storageCutoff = new Date()
  storageCutoff.setDate(storageCutoff.getDate() - storagePeriodDays)
  const dropdownChecks = filteredChecks.filter(
    (c) =>
      (c.status === 'success' && new Date(c.checkedAt) >= storageCutoff) ||
      c.id === checkIdParam
  )
  // Resolve active check: checkIdParam (if in current filter) → first dropdown → latest in filter.
  const paramInFilter = checkIdParam ? filteredChecks.find((c) => c.id === checkIdParam) : undefined
  const activeCheckId = paramInFilter?.id || dropdownChecks[0]?.id || filteredChecks[0]?.id || ''
  const activeCheckIndex = filteredChecks.findIndex((c) => c.id === activeCheckId)
  const activeCheck = filteredChecks[activeCheckIndex]

  // Group section checks by sectionId, sorted newest first within each group.
  const sectionChecksBySectionId = useMemo(() => {
    const map = new Map<string, Check[]>()
    for (const sc of sectionChecks) {
      if (!sc.sectionId) continue
      const arr = map.get(sc.sectionId) ?? []
      arr.push(sc)
      map.set(sc.sectionId, arr)
    }
    // Each group is already sorted DESC (flattened from parent checks sorted DESC)
    return map
  }, [sectionChecks])

  // When "All sections" is selected and the active check is a parent check,
  // gather its section children for display.
  // If the active parent has no section children, show the most recent per section.
  const activeSectionChecks = useMemo(() => {
    if (!activeCheck || activeCheck.sectionId || !hasSections) return []
    // Try linked via parentCheckId first.
    const linked = sectionChecks.filter((c) => c.parentCheckId === activeCheck.id)
    if (linked.length > 0) return linked
    // Fallback for old data without parentCheckId: find section checks near the parent's time.
    const parentTime = new Date(activeCheck.checkedAt).getTime()
    const nearby = sectionChecks.filter(
      (c) => Math.abs(new Date(c.checkedAt).getTime() - parentTime) < 5 * 60 * 1000
    )
    if (nearby.length > 0) return nearby
    // Last resort: pick the most recent section check per section so we always show something.
    const result: Check[] = []
    for (const [, arr] of sectionChecksBySectionId) {
      if (arr[0]) result.push(arr[0])
    }
    return result
  }, [activeCheck, sectionChecks, sectionChecksBySectionId, hasSections])

  // For a specific-section view, find the previous check with the same sectionId.
  const previousCheck = filteredChecks
    .slice(activeCheckIndex + 1)
    .find((c) => c.status === 'success' && !!c.screenshotUrl && c.sectionId === activeCheck?.sectionId)

  // Find previous section checks for "All sections" comparison.
  // For each active section check, find the next-oldest check with the same sectionId.
  const previousSectionChecks = useMemo(() => {
    if (!activeCheck || activeCheck.sectionId || !hasSections) return []
    const activeIds = new Set(activeSectionChecks.map((c) => c.id))
    const result: Check[] = []
    for (const sc of activeSectionChecks) {
      if (!sc.sectionId) continue
      const allForSection = sectionChecksBySectionId.get(sc.sectionId) ?? []
      const prev = allForSection.find(
        (c) => !activeIds.has(c.id) && c.status === 'success' && !!c.screenshotUrl
      )
      if (prev) result.push(prev)
    }
    return result
  }, [activeCheck, activeSectionChecks, sectionChecksBySectionId, hasSections])

  const activeSection = activeCheck?.sectionId
    ? sections.find((s) => s.id === activeCheck.sectionId)
    : undefined

  const activeSectionName = activeSection?.name

  useEffect(() => {
    async function loadData() {
      try {
        setLoading(true)
        const [checksData, pageData, usageData, sectionsData] = await Promise.all([
          ChangesViewService.getPageChecks(pageId),
          ChangesViewService.getPage(pageId),
          UsageApi.getChecksData(),
          PageApi.listSections(pageId),
        ])
        setStoragePeriodDays(usageData.storagePeriodDays)
        // Flatten: parent checks + their nested section checks into a single array.
        const allChecks: Check[] = []
        for (const check of checksData) {
          allChecks.push(check)
          if (check.sections) {
            allChecks.push(...check.sections)
          }
        }
        const sortedChecks = allChecks.sort(
          (a: Check, b: Check) => new Date(b.checkedAt).getTime() - new Date(a.checkedAt).getTime()
        )
        setChecks(sortedChecks)
        setPage(pageData)
        setSections(sectionsData)

        const detected = sortedChecks.filter((c: Check) => c.changeDetected)
        const resolvedCheckId = checkIdParam || detected[0]?.id
        if (resolvedCheckId) {
          const insightsData = await ChangesViewService.getPageInsights(pageId, resolvedCheckId)
          setInsights(insightsData)
        }
      } catch (error) {
        console.error('Failed to load changes data:', error)
      } finally {
        setLoading(false)
      }
    }
    loadData()
  }, [pageId, checkIdParam])

  // Build section-grouped text changes for the TextChanges component.
  const textChangeSections = useMemo<TextChangeSection[]>(() => {
    // "All sections" view: one group per section check (include all, even without text diffs)
    if (selectedSectionId === 'all' && activeSectionChecks.length > 0) {
      const sectionDiffs: TextChangeSection[] = activeSectionChecks.map((sc) => ({
        sectionName: sections.find((s) => s.id === sc.sectionId)?.name,
        changes: sc.contentDiff?.has_changes
          ? sc.contentDiff.diffs.map(blockDiffToDiffRow)
          : [],
        changeDetected: sc.changeDetected,
      }))
      // At least one section check exists — use section-grouped view.
      // If none have text diffs but parent has a full-page diff, fall through.
      if (sectionDiffs.some((s) => s.changes.length > 0 || s.changeDetected)) return sectionDiffs
    }
    // Single section or full-page: single group
    if (!activeCheck?.contentDiff?.has_changes) return []
    return [{
      sectionName: activeCheck.sectionId
        ? sections.find((s) => s.id === activeCheck.sectionId)?.name
        : undefined,
      changes: activeCheck.contentDiff.diffs.map(blockDiffToDiffRow),
    }]
  }, [activeCheck, selectedSectionId, activeSectionChecks, sections])

  if (loading) {
    return (
      <div className="flex-1 p-8 max-w-7xl mx-auto w-full">
        <div className="flex flex-col gap-6 md:gap-8 px-4 md:px-0">
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div className="flex flex-col gap-0.5">
              <span className="text-xs text-muted-foreground uppercase tracking-wide font-medium">Change detected on</span>
              <div className="h-9 w-56 bg-muted rounded animate-pulse" />
            </div>
            <div className="w-full md:w-64 flex flex-col gap-1">
              <div className="h-9 w-full bg-muted rounded-md animate-pulse" />
            </div>
          </div>

          <div className="border-b border-border overflow-x-auto">
            <div className="flex gap-6 md:gap-8 min-w-max">
              <span className="pb-3 text-sm font-medium border-b-2 border-primary text-foreground">
                Visual Pulse
              </span>
              <span className="pb-3 text-sm font-medium border-b-2 border-transparent text-muted-foreground">
                Text Changes
              </span>
              <span className="pb-3 text-sm font-medium border-b-2 border-transparent text-muted-foreground">
                Intelligent Insights
              </span>
            </div>
          </div>

          <div className="min-h-[500px]">
            <VisualPulseSkeleton />
          </div>
        </div>
      </div>
    )
  }

  // In "All sections" mode with a parent check active, render one VisualPulse per section.
  const renderAllSectionsView = () => {
    return (
      <div className="flex flex-col gap-8">
        {activeSectionChecks.map((sc) => {
          const sectionMeta = sections.find((s) => s.id === sc.sectionId)
          const prevSc = previousSectionChecks.find((p) => p.sectionId === sc.sectionId)
          return (
            <div key={sc.id}>
              {sectionMeta && (
                <h4 className="text-sm font-medium text-muted-foreground mb-2">{sectionMeta.name}</h4>
              )}
              <VisualPulse
                currentScreenshotUrl={sc.screenshotUrl}
                previousScreenshotUrl={prevSc?.screenshotUrl}
                sectionName={sectionMeta?.name}
              />
            </div>
          )
        })}
        {/* Full page screenshot as a bonus if the parent check has one */}
        {activeCheck?.screenshotUrl && (
          <div>
            <h4 className="text-sm font-medium text-muted-foreground mb-2">Full Page</h4>
            <VisualPulse
              currentScreenshotUrl={activeCheck.screenshotUrl}
              previousScreenshotUrl={previousCheck?.screenshotUrl}
            />
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="flex-1 p-8 max-w-7xl mx-auto w-full">
      {/* Section nav tabs — shown when page has sections */}
      {sections.length > 0 && (
        <div className="mb-6 border-b border-border pb-4">
          <SectionNavTabs
            sections={sections}
            selectedSectionId={selectedSectionId}
            onSelect={setSelectedSectionId}
            checks={checks}
            activeSectionChecks={activeSectionChecks}
          />
        </div>
      )}

      <ChangesViewLayout
        checks={dropdownChecks}
        activeCheckId={activeCheckId}
        activeTab={activeTab}
        onTabChange={setActiveTab}
        storagePeriodDays={storagePeriodDays}
        sections={sections}
      >
        {activeTab === 'visual' && (
          selectedSectionId === 'all' && activeSectionChecks.length > 0 ? (
            renderAllSectionsView()
          ) : (
            <VisualPulse
              currentScreenshotUrl={activeCheck?.screenshotUrl}
              previousScreenshotUrl={previousCheck?.screenshotUrl}
              sectionName={activeSectionName}
            />
          )
        )}
        {activeTab === 'text' && <TextChanges sections={textChangeSections} />}
        {activeTab === 'insights' && (
          <IntelligentInsights
            insights={insights}
            check={activeCheck}
            pageUrl={page?.url}
            screenshotUrl={activeCheck?.screenshotUrl}
            workspaceId={workspaceId}
            pageId={pageId}
            checkId={activeCheckId}
            onInsightsGenerated={setInsights}
          />
        )}
      </ChangesViewLayout>
    </div>
  )
}
