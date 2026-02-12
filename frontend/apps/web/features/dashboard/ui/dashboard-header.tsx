'use client'

import { Plus } from 'lucide-react'
import { StatCard } from './stat-card'
import type { DashboardStats } from '../domain/types'
import { Button } from '@workspace/ui/components/atoms'

export interface DashboardHeaderProps {
  userName: string
  stats: DashboardStats
  onCreateWorkspace: () => void
}

export function DashboardHeader({
  userName,
  stats,
  onCreateWorkspace,
}: Readonly<DashboardHeaderProps>) {
  return (
    <div className="bg-background px-4 md:px-8 lg:px-24 py-6 space-y-5">
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl md:text-4xl font-semibold text-foreground leading-tight">
            Hello {userName}!
          </h1>
          <p className="text-sm text-foreground/65 leading-snug">
            Your space to see how you're doing, what you've achieved, and what's next.
          </p>
        </div>
        <div className="flex gap-4">
          <Button
            onClick={onCreateWorkspace}
            className="bg-background hover:bg-muted text-foreground border border-border shadow-sm h-9 px-4 gap-2 flex-1 md:flex-none justify-center"
          >
            <Plus className="w-4 h-4" />
            Create workplace
          </Button>
          <Button
            onClick={onCreateWorkspace}
            className="bg-primary hover:bg-primary/90 text-primary-foreground shadow-sm h-9 px-4 gap-2 flex-1 md:flex-none justify-center"
          >
            <Plus className="w-4 h-4" />
            Add website
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          value={stats.workplaces.toString()}
          max={stats.maxWorkplaces.toString()}
          label="Workplaces"
        />
        <StatCard value={stats.pages.toString()} max={stats.maxPages.toString()} label="Pages" />
        <StatCard value={stats.todayChecks.toString()} label="Today's checks" />
        <StatCard
          value={stats.monthlyChecks.toString()}
          max={stats.maxMonthlyChecks.toString()}
          label="Monthly Checks"
          tag={`Usage ${stats.usagePercent}%`}
          tagColor="bg-accent text-accent-foreground border border-accent-foreground/20"
        />
      </div>
    </div>
  )
}
