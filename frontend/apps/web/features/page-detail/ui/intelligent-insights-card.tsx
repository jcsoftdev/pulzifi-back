'use client'

import type { Insight } from '@workspace/services/page-api'
import { Check } from 'lucide-react'

interface IntelligentInsightsCardProps {
  insights: Insight[]
}

export function IntelligentInsightsCard({ insights }: Readonly<IntelligentInsightsCardProps>) {
  return (
    <div className="flex flex-col gap-6 bg-card border border-border rounded-xl p-6 h-full">
      <h3 className="text-xl font-semibold text-foreground">Intelligent Insights</h3>

      <div className="flex flex-col gap-3">
        {insights.length === 0 ? (
          <p className="text-sm text-muted-foreground">No insights available yet.</p>
        ) : (
          insights.map((insight) => (
            <div key={insight.id} className="flex flex-col gap-1">
              <div className="flex items-center gap-2">
                <div className="flex items-center justify-center w-4 h-4 rounded-sm bg-primary">
                  <Check className="w-3 h-3 text-white" />
                </div>
                <span className="text-sm font-medium text-foreground">{insight.title}</span>
              </div>
              <p className="text-sm text-muted-foreground pl-6">{insight.content}</p>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
