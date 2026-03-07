'use client'

import type { Check, MonitoredSection } from '@workspace/services/page-api'
import { formatDateTime } from '@workspace/ui'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@workspace/ui/components/atoms/select'
import { cn } from '@workspace/ui/lib/utils'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import { useState } from 'react'
import { getSectionColor } from '@/features/page/domain/section-colors'

interface ChangesViewLayoutProps {
  checks: Check[]
  activeCheckId: string
  children: React.ReactNode
  activeTab?: string // 'visual' | 'text' | 'insights'
  onTabChange?: (tab: string) => void
  storagePeriodDays?: number
  sections?: MonitoredSection[]
}

export function ChangesViewLayout({
  checks,
  activeCheckId,
  children,
  activeTab: controlledActiveTab,
  onTabChange,
  storagePeriodDays,
  sections = [],
}: Readonly<ChangesViewLayoutProps>) {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  // Local state for tabs if not controlled
  const [internalActiveTab, setInternalActiveTab] = useState('visual')
  const activeTab = controlledActiveTab || internalActiveTab
  const handleTabChange = onTabChange || setInternalActiveTab

  const activeCheck = checks.find((c) => c.id === activeCheckId) || checks[0]
  const resolvedCheckId = activeCheck?.id || ''
  const activeCheckFailed =
    !!activeCheck && (activeCheck.status === 'error' || activeCheck.status === 'failed')

  const handleCheckChange = (checkId: string) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set('checkId', checkId)
    router.push(`${pathname}?${params.toString()}`)
  }

  // Build a section lookup for coloring check entries
  const sectionById = new Map(sections.map((s) => [s.id, s]))

  return (
    <div className="flex flex-col gap-6 md:gap-8 px-4 md:px-0">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div className="flex flex-col gap-0.5">
          <span className="text-xs text-muted-foreground uppercase tracking-wide font-medium">
            {activeCheckFailed ? 'Check failed on' : activeCheck?.changeDetected ? 'Change detected on' : 'Checked on'}
          </span>
          <h1 className="text-2xl md:text-3xl font-bold text-foreground">
            {activeCheck ? formatDateTime(activeCheck.checkedAt) : <span className="text-muted-foreground/40">—</span>}
          </h1>
          {activeCheck?.extractorFailed && (
            <span className="text-sm text-destructive mt-0.5">
              Extractor failed{activeCheck.errorMessage ? `: ${activeCheck.errorMessage}` : ''}
            </span>
          )}
        </div>

        <div className="w-full md:w-64 flex flex-col gap-1">
          {storagePeriodDays && (
            <span className="text-xs text-muted-foreground text-right">
              Storage period: {storagePeriodDays} days
            </span>
          )}
          <Select value={resolvedCheckId} onValueChange={handleCheckChange} disabled={checks.length === 0}>
            <SelectTrigger className="w-full bg-background">
              <SelectValue placeholder="No changes detected">
                {activeCheck ? formatDateTime(activeCheck.checkedAt) : null}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {checks.map((check) => {
                const section = check.sectionId ? sectionById.get(check.sectionId) : undefined
                const color = section ? getSectionColor(section.sortOrder) : undefined
                return (
                  <SelectItem key={check.id} value={check.id} textValue={formatDateTime(check.checkedAt)}>
                    <span className="flex items-center gap-2">
                      {color && (
                        <span
                          className="w-2 h-2 rounded-full shrink-0 inline-block"
                          style={{ backgroundColor: color }}
                        />
                      )}
                      <span>
                        {formatDateTime(check.checkedAt)}
                        {section ? ` — ${section.name}` : ''}
                        {check.extractorFailed ? ' — Extractor failed' : ''}
                      </span>
                    </span>
                  </SelectItem>
                )
              })}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-border overflow-x-auto">
        <div className="flex gap-6 md:gap-8 min-w-max">
          <button
            type="button"
            onClick={() => handleTabChange('visual')}
            className={cn(
              'pb-3 text-sm font-medium border-b-2 transition-colors',
              activeTab === 'visual'
                ? 'border-primary text-foreground'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            )}
          >
            Visual Pulse
          </button>
          <button
            type="button"
            onClick={() => handleTabChange('text')}
            className={cn(
              'pb-3 text-sm font-medium border-b-2 transition-colors',
              activeTab === 'text'
                ? 'border-primary text-foreground'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            )}
          >
            Text Changes
          </button>
          <button
            type="button"
            onClick={() => handleTabChange('insights')}
            className={cn(
              'pb-3 text-sm font-medium border-b-2 transition-colors',
              activeTab === 'insights'
                ? 'border-primary text-foreground'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            )}
          >
            Intelligent Insights
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="min-h-[500px]">{children}</div>
    </div>
  )
}
