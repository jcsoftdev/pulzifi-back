'use client'

import { DashboardHeader } from './ui/dashboard-header'
import type { DashboardStats } from './domain/types'

export function DashboardFeature() {
  const stats: DashboardStats = {
    workplaces: 3,
    maxWorkplaces: 10,
    pages: 20,
    maxPages: 200,
    todayChecks: 40,
    monthlyChecks: 100,
    maxMonthlyChecks: 2000,
    usagePercent: 5,
  }

  return (
    <div className="flex-1 flex flex-col">
      <DashboardHeader
        userName="Dania"
        stats={stats}
        onCreateWorkspace={() => console.log('Create workspace')}
      />

      <div className="px-24 py-8 bg-background">
        <div className="grid grid-cols-3 gap-6">
          {/* Found Changes Chart */}
          <div className="col-span-2 bg-card border border-border rounded-lg p-6">
            <div className="space-y-4">
              <div>
                <h3 className="text-xl font-semibold text-foreground mb-1">Found Changes</h3>
                <p className="text-sm text-muted-foreground">
                  Showing updates since each workspace was created.
                </p>
              </div>
              <div className="h-72 flex items-end gap-2">
                <div className="flex-1 flex flex-col items-center gap-2">
                  <div className="w-full h-36 bg-blue-700 rounded flex items-center justify-center">
                    <span className="text-primary-foreground text-sm font-semibold">Jeep</span>
                  </div>
                  <span className="text-sm font-semibold text-foreground">15</span>
                </div>
                <div className="flex-1 flex flex-col items-center gap-2">
                  <div className="w-full h-60 bg-cyan-700 rounded flex items-center justify-center">
                    <span className="text-primary-foreground text-sm font-semibold">Toyota</span>
                  </div>
                  <span className="text-sm font-semibold text-foreground">30</span>
                </div>
                <div className="flex-1 flex flex-col items-center gap-2">
                  <div className="w-full h-20 bg-orange-700 rounded flex items-center justify-center">
                    <span className="text-primary-foreground text-sm font-semibold">Nissan</span>
                  </div>
                  <span className="text-sm font-semibold text-foreground">5</span>
                </div>
              </div>
              <p className="text-xs text-muted-foreground text-center">Last scanning: 5 min ago</p>
            </div>
          </div>

          {/* AI Summary Insights */}
          <div className="bg-card border border-border rounded-lg p-6">
            <div className="space-y-4">
              <div>
                <h3 className="text-xl font-semibold text-foreground mb-1">AI Summary Insights</h3>
                <p className="text-sm text-muted-foreground">
                  Quick check of your last recommendations
                </p>
              </div>
              <div className="space-y-4 max-h-96 overflow-y-auto py-2">
                <div className="bg-card border border-border rounded-lg p-4 space-y-2">
                  <div className="flex justify-between items-start">
                    <span className="text-sm font-normal text-muted-foreground">toyota.com</span>
                    <span className="text-xs font-normal text-foreground">Today</span>
                  </div>
                  <p className="text-sm font-semibold text-foreground">/product</p>
                  <p className="text-xs font-normal text-foreground leading-snug">
                    After presenting 34 changes across different parts of the page's text, creating
                    three new sections about milestones...
                  </p>
                </div>

                <div className="bg-accent border border-accent-foreground rounded-lg p-4 space-y-2">
                  <div className="flex justify-between items-start">
                    <span className="text-sm font-normal text-muted-foreground">jeep.com</span>
                    <span className="text-xs font-normal text-foreground">Yesterday</span>
                  </div>
                  <p className="text-sm font-semibold text-foreground">/pricing</p>
                  <p className="text-xs font-normal text-foreground leading-snug">
                    After presenting 34 changes across different parts of the page's text, creating
                    three new sections about milestones...
                  </p>
                </div>

                <div className="bg-card border border-border rounded-lg p-4 space-y-2">
                  <div className="flex justify-between items-start">
                    <span className="text-sm font-normal text-muted-foreground">nissan.com</span>
                    <span className="text-xs font-normal text-foreground">Yesterday</span>
                  </div>
                  <p className="text-sm font-semibold text-foreground">/pricing</p>
                  <p className="text-xs font-normal text-foreground leading-snug">
                    After presenting 34 changes across different parts of the page's text, creating
                    three new sections about milestones...
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Recent Alerts Table */}
        <div className="mt-6 bg-card border border-border rounded-lg p-6">
          <div className="flex justify-between items-center mb-4">
            <div>
              <h3 className="text-xl font-semibold text-foreground mb-1">Recent Alerts</h3>
              <p className="text-sm text-muted-foreground">Quick overview about the last changes</p>
            </div>
            <button
              type="button"
              className="px-4 py-2 border border-border rounded-md text-sm text-foreground hover:bg-muted transition-colors"
            >
              View All
            </button>
          </div>

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
                <tr className="border-b border-border">
                  <td className="px-4 py-4 text-sm text-foreground">Today</td>
                  <td className="px-4 py-4 text-sm font-semibold text-foreground">Toyota</td>
                  <td className="px-4 py-4 text-sm text-foreground">Visual</td>
                  <td className="px-4 py-4 text-sm text-foreground">/servicios</td>
                </tr>
                <tr className="bg-secondary border-b border-border">
                  <td className="px-4 py-4 text-sm text-foreground">Today</td>
                  <td className="px-4 py-4 text-sm font-semibold text-foreground">Jeep</td>
                  <td className="px-4 py-4 text-sm text-foreground">Text</td>
                  <td className="px-4 py-4 text-sm text-foreground">/las-cosas-s...</td>
                </tr>
                <tr className="border-b border-border">
                  <td className="px-4 py-4 text-sm text-foreground">Yesterday</td>
                  <td className="px-4 py-4 text-sm font-semibold text-foreground">Nissan</td>
                  <td className="px-4 py-4 text-sm text-foreground">HTML</td>
                  <td className="px-4 py-4 text-sm text-foreground">/pricing</td>
                </tr>
                <tr className="border-b border-border">
                  <td className="px-4 py-4 text-sm text-foreground">Yesterday</td>
                  <td className="px-4 py-4 text-sm font-semibold text-foreground">Jeep</td>
                  <td className="px-4 py-4 text-sm text-foreground">HTML</td>
                  <td className="px-4 py-4 text-sm text-foreground">/</td>
                </tr>
                <tr>
                  <td className="px-4 py-4 text-sm text-foreground">Yesterday</td>
                  <td className="px-4 py-4 text-sm font-semibold text-foreground">Jeep</td>
                  <td className="px-4 py-4 text-sm text-foreground">Visual</td>
                  <td className="px-4 py-4 text-sm text-foreground">/pricing</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  )
}
