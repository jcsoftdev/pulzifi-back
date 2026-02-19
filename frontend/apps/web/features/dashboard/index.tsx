'use client'

import { useDashboardStats } from './application/use-dashboard-stats'
import type {
  DashboardStats,
  RecentAlert,
  RecentInsight,
  WorkspaceChanges,
} from './domain/types'
import { DashboardHeader } from './ui/dashboard-header'
import { EmptyStateCard } from './ui/empty-state-card'

function formatDate(isoString: string): string {
  const date = new Date(isoString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

  if (diffDays === 0) return 'Today'
  if (diffDays === 1) return 'Yesterday'
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}

function extractPath(url: string): string {
  try {
    return new URL(url).pathname || '/'
  } catch {
    return url
  }
}

function extractDomain(url: string): string {
  try {
    return new URL(url).hostname
  } catch {
    return url
  }
}

function ChangesChart({ workspaces }: { workspaces: WorkspaceChanges[] }) {
  const max = Math.max(...workspaces.map((w) => w.detectedChanges), 1)
  const colors = [
    'bg-blue-700',
    'bg-cyan-700',
    'bg-orange-700',
    'bg-violet-700',
    'bg-emerald-700',
  ]

  return (
    <div className="h-72 flex items-end gap-2">
      {workspaces.map((ws, i) => {
        const heightPercent = Math.max((ws.detectedChanges / max) * 100, 4)
        return (
          <div key={ws.workspaceName} className="flex-1 flex flex-col items-center gap-2">
            <div
              className={`w-full ${colors[i % colors.length]} rounded flex items-center justify-center`}
              style={{ height: `${heightPercent}%` }}
            >
              <span className="text-primary-foreground text-xs font-semibold truncate px-1">
                {ws.workspaceName}
              </span>
            </div>
            <span className="text-sm font-semibold text-foreground">{ws.detectedChanges}</span>
          </div>
        )
      })}
    </div>
  )
}

function InsightCard({ insight, highlighted }: { insight: RecentInsight; highlighted?: boolean }) {
  return (
    <div
      className={`${highlighted ? 'bg-accent border-accent-foreground' : 'bg-card border-border'} border rounded-lg p-4 space-y-2`}
    >
      <div className="flex justify-between items-start">
        <span className="text-sm font-normal text-muted-foreground">
          {extractDomain(insight.pageUrl)}
        </span>
        <span className="text-xs font-normal text-foreground">{formatDate(insight.createdAt)}</span>
      </div>
      <p className="text-sm font-semibold text-foreground">{extractPath(insight.pageUrl)}</p>
      <p className="text-xs font-normal text-foreground leading-snug line-clamp-3">
        {insight.content}
      </p>
    </div>
  )
}

function AlertsTable({ alerts }: { alerts: RecentAlert[] }) {
  return (
    <div className="border border-border rounded-lg overflow-hidden">
      <table className="w-full">
        <thead className="bg-muted border-b border-border">
          <tr>
            <th className="text-left px-4 py-3 text-xs font-semibold text-muted-foreground">
              Date
            </th>
            <th className="text-left px-4 py-3 text-xs font-semibold text-muted-foreground">
              Workplace
            </th>
            <th className="text-left px-4 py-3 text-xs font-semibold text-muted-foreground">
              Type
            </th>
            <th className="text-left px-4 py-3 text-xs font-semibold text-muted-foreground">
              Page
            </th>
          </tr>
        </thead>
        <tbody>
          {alerts.map((alert, i) => (
            <tr
              key={`${alert.checkedAt}-${alert.pageUrl}`}
              className={`${i % 2 === 1 ? 'bg-secondary' : ''} border-b border-border last:border-b-0`}
            >
              <td className="px-4 py-4 text-sm text-foreground">{formatDate(alert.checkedAt)}</td>
              <td className="px-4 py-4 text-sm font-semibold text-foreground">
                {alert.workspaceName}
              </td>
              <td className="px-4 py-4 text-sm text-foreground capitalize">{alert.changeType}</td>
              <td
                className="px-4 py-4 text-sm text-foreground truncate max-w-[160px]"
                title={alert.pageUrl}
              >
                {extractPath(alert.pageUrl)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function LoadingPlaceholder() {
  return (
    <div className="animate-pulse space-y-4">
      <div className="h-72 bg-muted rounded-lg" />
    </div>
  )
}

export function DashboardFeature() {
  const { stats, loading } = useDashboardStats()

  const headerStats: DashboardStats = {
    workplaces: stats?.workspacesCount ?? 0,
    maxWorkplaces: 10,
    pages: stats?.pagesCount ?? 0,
    maxPages: 200,
    todayChecks: stats?.todayChecksCount ?? 0,
    monthlyChecks: 0,
    maxMonthlyChecks: 2000,
    usagePercent: 0,
  }

  const isEmpty =
    !loading &&
    (!stats ||
      (stats.workspacesCount === 0 &&
        stats.pagesCount === 0 &&
        stats.recentInsights.length === 0))

  return (
    <div className="flex-1 flex flex-col">
      <DashboardHeader
        userName=""
        stats={headerStats}
        onCreateWorkspace={() => console.log('Create workspace')}
        onAddPage={() => console.log('Add new page')}
      />

      {isEmpty ? (
        <div className="px-4 md:px-8 lg:px-24 py-8 bg-background space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <EmptyStateCard
              title="No Workspaces Yet"
              description="Start tracking websites and monitoring changes by creating your first workspaces."
              buttonText="Create workspace"
              onButtonClick={() => console.log('Create workspace')}
            />
            <EmptyStateCard
              title="No Pages Yet"
              description="Add a page link to your workspace to start monitoring updates and tracking changes in real time."
              buttonText="+ Add new page"
              onButtonClick={() => console.log('Add new page')}
            />
          </div>
          <EmptyStateCard
            title="No Insights Yet"
            description="You'll see AI insights once an alert is detected."
            buttonText="Go to Settings"
            onButtonClick={() => console.log('Go to settings')}
          />
        </div>
      ) : (
        <div className="px-4 md:px-8 lg:px-24 py-8 bg-background">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Found Changes Chart */}
            <div className="col-span-1 lg:col-span-2 bg-card border border-border rounded-lg p-6">
              <div className="space-y-4">
                <div>
                  <h3 className="text-xl font-semibold text-foreground mb-1">Found Changes</h3>
                  <p className="text-sm text-muted-foreground">
                    Showing updates since each workspace was created.
                  </p>
                </div>

                {loading ? (
                  <LoadingPlaceholder />
                ) : stats && stats.changesPerWorkspace.length > 0 ? (
                  <ChangesChart workspaces={stats.changesPerWorkspace} />
                ) : (
                  <div className="h-72 flex items-center justify-center text-muted-foreground text-sm">
                    No workspaces with changes yet.
                  </div>
                )}
              </div>
            </div>

            {/* AI Summary Insights */}
            <div className="bg-card border border-border rounded-lg p-6">
              <div className="space-y-4">
                <div>
                  <h3 className="text-xl font-semibold text-foreground mb-1">
                    AI Summary Insights
                  </h3>
                  <p className="text-sm text-muted-foreground">
                    Quick check of your last recommendations
                  </p>
                </div>
                <div className="space-y-4 max-h-96 overflow-y-auto py-2">
                  {loading ? (
                    <div className="animate-pulse space-y-3">
                      {[1, 2, 3].map((n) => (
                        <div key={n} className="h-24 bg-muted rounded-lg" />
                      ))}
                    </div>
                  ) : stats && stats.recentInsights.length > 0 ? (
                    stats.recentInsights.map((insight, i) => (
                      <InsightCard
                        key={insight.createdAt + insight.pageUrl}
                        insight={insight}
                        highlighted={i === 1}
                      />
                    ))
                  ) : (
                    <p className="text-sm text-muted-foreground">No insights generated yet.</p>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Recent Alerts Table */}
          <div className="mt-6 bg-card border border-border rounded-lg p-6">
            <div className="flex justify-between items-center mb-4">
              <div>
                <h3 className="text-xl font-semibold text-foreground mb-1">Recent Alerts</h3>
                <p className="text-sm text-muted-foreground">
                  Quick overview about the last changes
                </p>
              </div>
              <button
                type="button"
                className="px-4 py-2 border border-border rounded-md text-sm text-foreground hover:bg-muted transition-colors"
              >
                View All
              </button>
            </div>

            {loading ? (
              <div className="animate-pulse space-y-2">
                {[1, 2, 3, 4].map((n) => (
                  <div key={n} className="h-12 bg-muted rounded" />
                ))}
              </div>
            ) : stats && stats.recentAlerts.length > 0 ? (
              <AlertsTable alerts={stats.recentAlerts} />
            ) : (
              <p className="text-sm text-muted-foreground py-4 text-center">
                No changes detected yet.
              </p>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
