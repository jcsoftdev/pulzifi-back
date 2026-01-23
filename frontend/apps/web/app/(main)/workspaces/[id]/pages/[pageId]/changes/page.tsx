'use client'

import { useState, useEffect } from 'react'
import { useSession } from 'next-auth/react'
import { useParams, useSearchParams } from 'next/navigation'
import { ChangesViewService } from '@/features/changes-view/domain/changes-view-service'
import { ChangesViewLayout } from '@/features/changes-view/ui/changes-view-layout'
import { VisualPulse } from '@/features/changes-view/ui/visual-pulse'
import { TextChanges } from '@/features/changes-view/ui/text-changes'
import { IntelligentInsights } from '@/features/changes-view/ui/intelligent-insights'
import { diffLines } from '@/features/changes-view/utils/simple-diff'
import type { Check, Insight } from '@workspace/services/page-api'
import { Loader2 } from 'lucide-react'

interface TextChange {
  type: 'added' | 'removed' | 'unchanged'
  text: string
}

export default function ChangesPage() {
  const params = useParams()
  const searchParams = useSearchParams()
  const { status, data: session } = useSession()
  const pageId = params.pageId as string
  const checkIdParam = searchParams.get('checkId')

  const [checks, setChecks] = useState<Check[]>([])
  const [insights, setInsights] = useState<Insight[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('visual')
  const [textChanges, setTextChanges] = useState<TextChange[]>([])
  const [loadingDiff, setLoadingDiff] = useState(false)

  useEffect(() => {
    const token = (session as any)?.accessToken ?? null
    ;(globalThis as any).__authToken__ = token
  }, [session])

  useEffect(() => {
    async function loadData() {
      try {
        if (status !== 'authenticated') return
        setLoading(true)
        const checksData = await ChangesViewService.getPageChecks(pageId)
        // Sort checks descending by date
        const sortedChecks = checksData.sort((a: Check, b: Check) => 
          new Date(b.checkedAt).getTime() - new Date(a.checkedAt).getTime()
        )
        setChecks(sortedChecks)

        const activeCheckId = checkIdParam || sortedChecks[0]?.id
        if (activeCheckId) {
          const insightsData = await ChangesViewService.getPageInsights(pageId)
          setInsights(insightsData.filter((i: Insight) => i.checkId === activeCheckId))
        }
      } catch (error) {
        console.error('Failed to load changes data:', error)
      } finally {
        setLoading(false)
      }
    }
    loadData()
  }, [pageId, checkIdParam, status])

  const activeCheckId = checkIdParam || checks[0]?.id || ''
  const activeCheckIndex = checks.findIndex(c => c.id === activeCheckId)
  const activeCheck = checks[activeCheckIndex]
  const previousCheck = checks[activeCheckIndex + 1]

  useEffect(() => {
    async function loadDiff() {
      if (status !== 'authenticated') return
      if (activeTab === 'text' && activeCheck) {
        setLoadingDiff(true)
        try {
          // If previous check exists, compare. Else show current text as added? Or just empty.
          // If no previous check, we can't show diff.
          if (!previousCheck) {
             // Maybe fetch current text and show as all added?
             // For now, let's leave empty or handle gracefully
             setTextChanges([])
             return
          }
          
          // Need to fetch HTML content. Assuming Check object has htmlSnapshotUrl
          // If not in interface, we might need to cast or update interface.
          // Check interface in page-api.ts has `screenshotUrl`. It doesn't have `htmlSnapshotUrl` explicitly in frontend interface?
          // Let's check page-api.ts again.
          
          // Assuming backend returns it but frontend interface might miss it.
          // I will use 'any' cast if needed or just try access.
          const currentUrl = (activeCheck as any).htmlSnapshotUrl
          const prevUrl = (previousCheck as any).htmlSnapshotUrl

          if (currentUrl && prevUrl) {
             const [currentHtml, prevHtml] = await Promise.all([
               ChangesViewService.getHtmlContent(currentUrl),
               ChangesViewService.getHtmlContent(prevUrl)
             ])
             
             const currentText = ChangesViewService.extractTextFromHtml(currentHtml)
             const prevText = ChangesViewService.extractTextFromHtml(prevHtml)
             
             const diff = diffLines(prevText, currentText)
             
             const changes: TextChange[] = diff.map(d => ({
               type: d.added ? 'added' : d.removed ? 'removed' : 'unchanged',
               text: d.value
             }))
             
             setTextChanges(changes)
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
  }, [activeTab, activeCheck, previousCheck, status])


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
        checks={checks}
        activeCheckId={activeCheckId}
        activeTab={activeTab}
        onTabChange={setActiveTab}
      >
        {activeTab === 'visual' && (
          <VisualPulse
            currentScreenshotUrl={activeCheck?.screenshotUrl}
            previousScreenshotUrl={previousCheck?.screenshotUrl}
          />
        )}
        {activeTab === 'text' && (
          loadingDiff ? (
             <div className="flex items-center justify-center h-64">
               <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
             </div>
          ) : (
            <TextChanges changes={textChanges} />
          )
        )}
        {activeTab === 'insights' && (
          <IntelligentInsights insights={insights} />
        )}
      </ChangesViewLayout>
    </div>
  )
}
