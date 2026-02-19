'use client'

import type { Check, Insight, Page } from '@workspace/services/page-api'
import { UsageApi } from '@workspace/services'
import { Loader2 } from 'lucide-react'
import { useParams, useSearchParams } from 'next/navigation'
import { useEffect, useState } from 'react'
import { ChangesViewService } from '@/features/changes-view/domain/changes-view-service'
import { ChangesViewLayout } from '@/features/changes-view/ui/changes-view-layout'
import { IntelligentInsights } from '@/features/changes-view/ui/intelligent-insights'
import { TextChanges } from '@/features/changes-view/ui/text-changes'
import { VisualPulse } from '@/features/changes-view/ui/visual-pulse'
import type { DiffRow } from '@/features/changes-view/utils/simple-diff'
import { diffLines } from '@/features/changes-view/utils/simple-diff'

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

  // Only detected-change checks within the storage period appear in the dropdown
  const storageCutoff = new Date()
  storageCutoff.setDate(storageCutoff.getDate() - storagePeriodDays)
  const detectedChecks = checks.filter(
    (c) => c.changeDetected && new Date(c.checkedAt) >= storageCutoff
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
      <div className="flex items-center justify-center h-full min-h-[500px]">
        <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
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
            <div className="flex items-center justify-center h-64">
              <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
            </div>
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
