'use client'

import type { Check, Insight, Page } from '@workspace/services/page-api'
import { UsageApi } from '@workspace/services'
import { Button } from '@workspace/ui/components/atoms/button'
import { Settings, Zap } from 'lucide-react'
import { useParams, useSearchParams } from 'next/navigation'
import { useEffect, useState } from 'react'
import { ChangesViewService } from '@/features/changes-view/domain/changes-view-service'
import { ChangesViewLayout } from '@/features/changes-view/ui/changes-view-layout'
import { IntelligentInsights } from '@/features/changes-view/ui/intelligent-insights'
import { TextChanges } from '@/features/changes-view/ui/text-changes'
import { VisualPulse } from '@/features/changes-view/ui/visual-pulse'
import type { DiffRow } from '@/features/changes-view/utils/simple-diff'
import { diffLines } from '@/features/changes-view/utils/simple-diff'

function VisualPulseSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <div className="relative w-full select-none overflow-hidden rounded-lg border border-border shadow-sm bg-muted/10 h-[500px] animate-pulse">
        {/* Slider handle placeholder */}
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

function TextChangesSkeleton() {
  return (
    <div className="rounded-xl border border-border bg-card">
      <div className="px-5 py-3.5 border-b border-border">
        <h3 className="text-sm font-medium text-muted-foreground tracking-wide">Text Changes</h3>
      </div>
      <div className="p-4 space-y-2">
        {['a', 'b', 'c', 'd'].map((i) => (
          <div key={i} className="rounded-lg border border-border/40 overflow-hidden text-sm">
            <div className="flex gap-3 px-4 py-2.5 border-b border-border/30 bg-muted/10">
              <span className="select-none shrink-0 text-foreground/25 font-mono text-xs pt-px">−</span>
              <div className="h-4 w-full bg-muted rounded animate-pulse" />
            </div>
            <div className="flex gap-3 px-4 py-2.5 bg-emerald-950/20">
              <span className="select-none shrink-0 text-emerald-500/60 font-mono text-xs pt-px">+</span>
              <div className="h-4 w-full bg-muted rounded animate-pulse" />
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function InsightsSkeleton() {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <div className="lg:col-span-2 flex flex-col gap-6">
        {/* Empty state card matching real UI */}
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
            {/* Screenshot placeholder */}
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
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('visual')
  const [textChanges, setTextChanges] = useState<DiffRow[]>([])
  const [loadingDiff, setLoadingDiff] = useState(false)
  const [storagePeriodDays, setStoragePeriodDays] = useState(7)

  // Detected-change checks within the storage period appear in the dropdown.
  // The check referenced by the URL param is always included so the page
  // works even when that check is older than the storage period.
  const storageCutoff = new Date()
  storageCutoff.setDate(storageCutoff.getDate() - storagePeriodDays)
  const detectedChecks = checks.filter(
    (c) =>
      (c.changeDetected && new Date(c.checkedAt) >= storageCutoff) ||
      c.id === checkIdParam
  )
  const activeCheckId = checkIdParam || detectedChecks[0]?.id || ''
  // Find position in the FULL sorted list so previousCheck is always the
  // immediately preceding run (not the previous detected change)
  const activeCheckIndex = checks.findIndex((c) => c.id === activeCheckId)
  const activeCheck = checks[activeCheckIndex]
  const previousCheck = checks[activeCheckIndex + 1]

  useEffect(() => {
    async function loadData() {
      try {
        setLoading(true)
        const [checksData, pageData, usageData] = await Promise.all([
          ChangesViewService.getPageChecks(pageId),
          ChangesViewService.getPage(pageId),
          UsageApi.getChecksData(),
        ])
        setStoragePeriodDays(usageData.storagePeriodDays)
        // Sort checks descending by date
        const sortedChecks = checksData.sort(
          (a: Check, b: Check) => new Date(b.checkedAt).getTime() - new Date(a.checkedAt).getTime()
        )
        setChecks(sortedChecks)
        setPage(pageData)

        // Determine which check is active to fetch insights for it
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

  useEffect(() => {
    async function loadDiff() {
      if (activeTab === 'text' && activeCheck) {
        setLoadingDiff(true)
        try {
          if (!previousCheck) {
            setTextChanges([])
            return
          }

          const currentUrl = activeCheck.htmlSnapshotUrl
          const prevUrl = previousCheck.htmlSnapshotUrl

          if (currentUrl && prevUrl) {
            const [currentHtml, prevHtml] = await Promise.all([
              ChangesViewService.getHtmlContent(currentUrl),
              ChangesViewService.getHtmlContent(prevUrl),
            ])

            const currentText = ChangesViewService.extractTextFromHtml(currentHtml)
            const prevText = ChangesViewService.extractTextFromHtml(prevHtml)

            const diff = diffLines(prevText, currentText)

            setTextChanges(diff)
          } else {
            setTextChanges([])
          }
        } catch (error) {
          console.error('Failed to calculate diff:', error)
          setTextChanges([])
        } finally {
          setLoadingDiff(false)
        }
      }
    }
    loadDiff()
  }, [activeTab, activeCheck, previousCheck])

  if (loading) {
    return (
      <div className="flex-1 p-8 max-w-7xl mx-auto w-full">
        <div className="flex flex-col gap-6 md:gap-8 px-4 md:px-0">
          {/* Header */}
          <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
            <div className="flex flex-col gap-1">
              <span className="text-sm text-muted-foreground">Change detected on:</span>
              <div className="h-8 w-64 bg-muted rounded animate-pulse" />
            </div>
            <div className="w-full md:w-64 flex flex-col gap-1">
              <div className="h-9 w-full bg-muted rounded-md animate-pulse" />
            </div>
          </div>

          {/* Tabs — static */}
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

          {/* Content — Visual Pulse skeleton */}
          <div className="min-h-[500px]">
            <VisualPulseSkeleton />
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex-1 p-8 max-w-7xl mx-auto w-full">
      <ChangesViewLayout
        checks={detectedChecks}
        activeCheckId={activeCheckId}
        activeTab={activeTab}
        onTabChange={setActiveTab}
        storagePeriodDays={storagePeriodDays}
      >
        {activeTab === 'visual' && (
          <VisualPulse
            currentScreenshotUrl={activeCheck?.screenshotUrl}
            previousScreenshotUrl={previousCheck?.screenshotUrl}
          />
        )}
        {activeTab === 'text' &&
          (loadingDiff ? (
            <TextChangesSkeleton />
          ) : (
            <TextChanges changes={textChanges} />
          ))}
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
