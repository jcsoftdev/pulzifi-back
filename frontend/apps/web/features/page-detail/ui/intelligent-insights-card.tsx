'use client'

import { type MonitoringConfig, PageApi } from '@workspace/services/page-api'
import { Check } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { useTransition } from 'react'

const INSIGHT_TYPES = [
  { id: 'marketing', label: 'Marketing Lens' },
  { id: 'market_analysis', label: 'Market Analysis' },
  { id: 'business_opportunities', label: 'Business Opportunities' },
  { id: 'job_recommendation', label: 'Job recommendation' },
] as const

const ALERT_CONDITIONS = [
  { id: 'any_changes', label: 'Any changes' },
  { id: 'new_article', label: 'A new article is published on the site' },
  { id: 'new_comment', label: 'New comment added' },
  { id: 'main_nav_changes', label: "Site's main navigation menu changes" },
] as const

interface IntelligentInsightsCardProps {
  pageId: string
  config: MonitoringConfig | null
}

export function IntelligentInsightsCard({
  pageId,
  config,
}: Readonly<IntelligentInsightsCardProps>) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()

  const enabledInsightTypes = config?.enabledInsightTypes ?? ['marketing', 'market_analysis']
  const enabledAlertConditions = config?.enabledAlertConditions ?? ['any_changes']
  const customAlertCondition = config?.customAlertCondition ?? ''

  const save = (updates: Partial<MonitoringConfig>) => {
    startTransition(async () => {
      try {
        await PageApi.updateMonitoringConfig(pageId, updates)
        router.refresh()
      } catch (error) {
        console.error('Failed to update insight preferences', error)
      }
    })
  }

  const toggleInsightType = (id: string) => {
    const next = enabledInsightTypes.includes(id)
      ? enabledInsightTypes.filter((t) => t !== id)
      : [...enabledInsightTypes, id]
    save({ enabledInsightTypes: next })
  }

  const toggleAlertCondition = (id: string) => {
    const next = enabledAlertConditions.includes(id)
      ? enabledAlertConditions.filter((a) => a !== id)
      : [...enabledAlertConditions, id]
    save({ enabledAlertConditions: next })
  }

  const handleCustomAlertBlur = (value: string) => {
    if (value !== customAlertCondition) {
      save({ customAlertCondition: value })
    }
  }

  return (
    <div
      className={`flex flex-col gap-5 bg-card border border-primary/25 rounded-xl p-6 transition-opacity ${isPending ? 'opacity-70' : ''}`}
    >
      <h3 className="text-xl font-semibold text-foreground">Intelligent Insights</h3>

      {/* Insight types */}
      <div className="flex flex-col gap-3">
        {INSIGHT_TYPES.map((type) => {
          const checked = enabledInsightTypes.includes(type.id)
          return (
            <button
              key={type.id}
              type="button"
              disabled={isPending}
              className="flex items-center gap-3 cursor-pointer group text-left disabled:cursor-not-allowed"
              onClick={() => toggleInsightType(type.id)}
            >
              <div
                className={`flex items-center justify-center w-4 h-4 rounded-sm border flex-shrink-0 transition-colors ${
                  checked ? 'bg-primary border-primary' : 'border-muted-foreground'
                }`}
              >
                {checked && <Check className="w-3 h-3 text-primary-foreground" />}
              </div>
              <span className="text-sm text-foreground group-hover:text-foreground/80 transition-colors">
                {type.label}
              </span>
            </button>
          )
        })}
        <p className="text-sm text-muted-foreground pl-7">New insights coming soon...</p>
      </div>

      <div className="border-t border-border" />

      {/* Alert conditions */}
      <div className="flex flex-col gap-3">
        <p className="text-sm text-muted-foreground">Alert me when</p>

        {ALERT_CONDITIONS.map((condition) => {
          const checked = enabledAlertConditions.includes(condition.id)
          return (
            <button
              key={condition.id}
              type="button"
              disabled={isPending}
              className="flex items-center gap-3 cursor-pointer group text-left disabled:cursor-not-allowed"
              onClick={() => toggleAlertCondition(condition.id)}
            >
              <div
                className={`flex items-center justify-center w-4 h-4 rounded-sm border flex-shrink-0 transition-colors ${
                  checked ? 'bg-primary border-primary' : 'border-muted-foreground'
                }`}
              >
                {checked && <Check className="w-3 h-3 text-primary-foreground" />}
              </div>
              <span className="text-sm text-foreground group-hover:text-foreground/80 transition-colors">
                {condition.label}
              </span>
            </button>
          )
        })}

        <textarea
          rows={3}
          placeholder="Add your own here"
          defaultValue={customAlertCondition}
          onBlur={(e) => handleCustomAlertBlur(e.target.value)}
          disabled={isPending}
          className="mt-1 w-full resize-none rounded-lg border border-border bg-muted/30 px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary disabled:opacity-50"
        />
      </div>
    </div>
  )
}
