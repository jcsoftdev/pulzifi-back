'use client'

import { Plus } from "lucide-react"
import { StatCard } from "./stat-card"
import type { DashboardStats } from "../domain/types"
import { Button } from "@workspace/ui/components/atoms"

export interface DashboardHeaderProps {
  userName: string
  stats: DashboardStats
  onCreateWorkspace: () => void
}

export function DashboardHeader({ userName, stats, onCreateWorkspace }: Readonly<DashboardHeaderProps>) {
  return (
    <div className="bg-background px-24 py-6 space-y-5">
      <div className="flex items-end justify-between">
        <div className="space-y-2">
          <h1 className="text-[40px] font-semibold text-foreground leading-tight">
            Hello {userName}!
          </h1>
          <p className="text-[14.6px] text-foreground/65 leading-snug">
            Your space to see how you're doing, what you've achieved, and what's next.
          </p>
        </div>
        <div className="flex gap-4">
          <Button 
            onClick={onCreateWorkspace} 
            className="bg-background hover:bg-muted text-foreground border border-border shadow-[0px_2px_0px_0px_rgba(4,25,255,0.04)] h-[42px] px-4 gap-2"
          >
            <Plus className="w-4 h-4" />
            Create workplace
          </Button>
          <Button 
            onClick={onCreateWorkspace} 
            className="bg-primary hover:bg-primary/90 text-primary-foreground shadow-[0px_2px_0px_0px_rgba(5,145,255,0.1)] h-[42px] px-4 gap-2"
          >
            <Plus className="w-4 h-4" />
            Add website
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-4 gap-6">
        <StatCard
          value={stats.workplaces.toString()}
          max={stats.maxWorkplaces.toString()}
          label="Workplaces"
        />
        <StatCard
          value={stats.pages.toString()}
          max={stats.maxPages.toString()}
          label="Pages"
        />
        <StatCard
          value={stats.todayChecks.toString()}
          label="Today's checks"
        />
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
