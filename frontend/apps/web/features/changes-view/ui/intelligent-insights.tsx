'use client'

import type { Insight } from '@workspace/services/page-api'
import { Button } from '@workspace/ui/components/atoms/button'
import { Copy } from 'lucide-react'

interface IntelligentInsightsProps {
  insights: Insight[]
}

export function IntelligentInsights({ insights }: Readonly<IntelligentInsightsProps>) {
  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text)
  }

  // Group insights by type if needed, or just display them
  // The design shows sections like "Overview", "Marketing Lens", "Market analysis"
  // We'll assume the insights have a 'type' or 'title' that maps to these, or we render them as cards.
  // Based on the screenshot, it looks like specific sections.
  // If we don't have types, we'll just render them in order.

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <div className="lg:col-span-2 flex flex-col gap-6">
        {insights.length === 0 ? (
          <div className="bg-card border border-border rounded-xl p-8 text-center text-muted-foreground">
            No insights available for this check.
          </div>
        ) : (
          insights.map((insight) => (
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
          ))
        )}
      </div>

      <div className="lg:col-span-1">
        {/* Placeholder for sidebar or just empty as per design having a sidebar-like structure on the right in some views, 
             but the screenshot shows "Edit Intelligent Insights" button on the right */}
        <div className="flex flex-col gap-6">
          <Button className="w-full gap-2" variant="outline">
            <svg
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
              aria-hidden="true"
              focusable="false"
            >
              <path
                d="M8 1.5C4.41015 1.5 1.5 4.41015 1.5 8C1.5 11.5899 4.41015 14.5 8 14.5C11.5899 14.5 14.5 11.5899 14.5 8C14.5 4.41015 11.5899 1.5 8 1.5ZM8 1.5V14.5"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
            Edit Intelligent Insights
          </Button>

          <div className="bg-muted/30 rounded-lg p-6 flex flex-col items-center justify-center text-center gap-4 border border-border border-dashed">
            <p className="text-sm font-medium">Changes detected on June 15, 2025</p>
            <Button variant="outline" className="bg-background">
              View Site
            </Button>
            <p className="text-xs text-muted-foreground">www.toyota.com/aboutus</p>
          </div>
        </div>
      </div>
    </div>
  )
}
