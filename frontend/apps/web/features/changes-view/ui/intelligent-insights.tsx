'use client'

import type { Check, Insight } from '@workspace/services/page-api'
import { PageApi } from '@workspace/services/page-api'
import { Button } from '@workspace/ui/components/atoms/button'
import { formatDateTime } from '@workspace/ui'
import { Copy, Loader2, Settings, Zap } from 'lucide-react'
import Image from 'next/image'
import Link from 'next/link'
import { useState } from 'react'

interface IntelligentInsightsProps {
  insights: Insight[]
  check?: Check
  pageUrl?: string
  screenshotUrl?: string
  workspaceId?: string
  pageId?: string
  checkId?: string
  onInsightsGenerated?: (insights: Insight[]) => void
}

export function IntelligentInsights({
  insights: initialInsights,
  check,
  pageUrl,
  screenshotUrl,
  workspaceId,
  pageId,
  checkId,
  onInsightsGenerated,
}: Readonly<IntelligentInsightsProps>) {
  const sortInsights = (list: Insight[]) =>
    [...list].sort((a, b) => {
      if (a.insightType === 'overview') return -1
      if (b.insightType === 'overview') return 1
      return 0
    })

  const [insights, setInsights] = useState<Insight[]>(sortInsights(initialInsights))
  const [generating, setGenerating] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const parseInsightData = (rawData: unknown) => {
    const data = rawData as {
      timeout?: boolean
      insights?: Array<{
        id: string
        page_id: string
        check_id: string
        insight_type: string
        title: string
        content: string
        metadata: Record<string, unknown>
        created_at: string
      }>
    }
    if (data.timeout || !data.insights) {
      return []
    }
    return data.insights.map((i) => ({
      id: i.id,
      pageId: i.page_id,
      checkId: i.check_id,
      insightType: i.insight_type,
      title: i.title,
      content: i.content,
      metadata: i.metadata,
      createdAt: i.created_at,
    }))
  }

  const waitForInsightsSSE = (cid: string): Promise<Insight[]> =>
    new Promise((resolve, reject) => {
      const { protocol, host } = globalThis.location
      const url = `${protocol}//${host}/api/v1/insights/sse?check_id=${cid}`
      const source = new EventSource(url, { withCredentials: true })

      source.onmessage = (ev) => {
        source.close()
        try {
          const insights = parseInsightData(JSON.parse(ev.data))
          resolve(insights)
        } catch {
          reject(new Error('invalid SSE payload'))
        }
      }

      source.onerror = () => {
        source.close()
        reject(new Error('SSE connection error'))
      }
    })

  const handleCopy = (text: string) => {
    if (navigator.clipboard) {
      navigator.clipboard.writeText(text)
    } else {
      const el = document.createElement('textarea')
      el.value = text
      el.style.position = 'fixed'
      el.style.opacity = '0'
      document.body.appendChild(el)
      el.select()
      document.execCommand('copy')
      document.body.removeChild(el)
    }
  }

  const handleGenerate = async () => {
    if (!pageId || !checkId) return
    setGenerating(true)
    setError(null)
    try {
      // Trigger generation — backend returns 202 immediately and runs async.
      await PageApi.generateInsights(pageId, checkId)
      // Subscribe via SSE and wait for the broker to push the ready event.
      const results = await waitForInsightsSSE(checkId)
      if (results.length > 0) {
        setInsights(sortInsights(results))
        onInsightsGenerated?.(results)
      } else {
        setError('Insights are still being generated. Please refresh the page.')
      }
    } catch {
      setError('Failed to generate insights. Please try again.')
    } finally {
      setGenerating(false)
    }
  }

  const editHref =
    workspaceId && pageId ? `/workspaces/${workspaceId}/pages/${pageId}` : undefined

  const renderEmptyState = () => (
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
      {error && <p className="text-sm text-destructive">{error}</p>}
      {pageId && checkId ? (
        <Button
          className="mt-2 gap-2 bg-violet-600 hover:bg-violet-700 text-white"
          onClick={handleGenerate}
          disabled={generating}
        >
          {generating ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <Zap className="h-4 w-4" />
          )}
          {generating ? 'Generating…' : 'View Intelligent Insights'}
        </Button>
      ) : (
        <Button className="mt-2 gap-2 bg-violet-600 hover:bg-violet-700 text-white" disabled>
          <Zap className="h-4 w-4" />
          View Intelligent Insights
        </Button>
      )}
    </div>
  )

  const renderInsightsList = () => (
    <>
      {insights.map((insight) => (
        <div
          key={insight.id}
          className="bg-card border border-border rounded-xl p-6 flex flex-col gap-4"
        >
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold text-foreground">{insight.title}</h3>
            <Button
              variant="outline"
              size="sm"
              className="h-8 gap-2"
              onClick={() => handleCopy(insight.content)}
            >
              <Copy className="h-3.5 w-3.5" />
              <span className="text-xs">copy</span>
            </Button>
          </div>
          <div className="text-sm text-muted-foreground leading-relaxed whitespace-pre-wrap">
            {insight.content}
          </div>
        </div>
      ))}
    </>
  )

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <div className="lg:col-span-2 flex flex-col gap-6">
        {insights.length === 0 ? renderEmptyState() : renderInsightsList()}
      </div>

      <div className="lg:col-span-1">
        <div className="flex flex-col gap-6">
          
          {editHref ? (
            <Button asChild className="w-full gap-2" variant="outline">
              <Link href={editHref}>
                <Settings className="h-4 w-4" />
                Edit Intelligent Insights
              </Link>
            </Button>
          ) : (
            <Button className="w-full gap-2" variant="outline" disabled>
              <Settings className="h-4 w-4" />
              Edit Intelligent Insights
            </Button>
          )}
          <div className="bg-muted/30 rounded-lg p-6 flex flex-col items-center justify-center text-center gap-4 border border-border border-dashed">
            {screenshotUrl && (
              <div className="relative w-full aspect-[16/9] rounded-xl overflow-hidden bg-muted shadow-md border border-border">
                <Image
                  src={screenshotUrl}
                  alt="Page screenshot"
                  fill
                  className="object-cover object-top"
                  unoptimized
                />
              </div>
            )}
            <p className="text-sm font-medium">
              {check ? `Changes detected on ${formatDateTime(check.checkedAt)}` : 'No check selected'}
            </p>
            {pageUrl && (
              <Button variant="outline" className="bg-background" asChild>
                <a href={pageUrl} target="_blank" rel="noopener noreferrer">
                  View Site
                </a>
              </Button>
            )}
            {pageUrl && (
              <p className="text-xs text-muted-foreground break-all">{pageUrl}</p>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
